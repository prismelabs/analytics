package sessionstorage

import (
	"container/heap"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/negrel/assert"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// Service define an in-memory session storage.
type Service interface {
	// InsertSession stores given session in memory. If number of visitor session
	// exceed configured max session per visitor, this function returns false and
	// session isn't stored.
	InsertSession(deviceId uint64, session event.Session) bool
	// AddPageview adds a pageview to a session with the given device id
	// latest path (referrer). Session is returned along a true flag if it was
	// found.
	AddPageview(deviceId uint64, referrer event.ReferrerUri, uri uri.Uri) (event.Session, bool)
	// IdentifySession updates stored session visitor id. Updated session and
	// boolean found flag are returned.
	IdentifySession(deviceId uint64, pageUri uri.Uri, visitorId string) (event.Session, bool)
	// WaitSession retrieves stored session and returns it. If session is not
	// found, it waits until it is created or timeout.
	// Returned boolean flag is false if wait timed out and returned an empty
	// session.
	WaitSession(deviceId uint64, pageUri uri.Uri, timeout time.Duration) (event.Session, bool)
}

// sessionEntry holds session and associated metadata of an entry in session
// storage.
type sessionEntry struct {
	Session   event.Session
	latestUri uri.Uri
	wait      chan struct{}
	expiry    uint32
}

func (e *sessionEntry) hasWaiter() bool {
	return e.wait != nil
}

func (e *sessionEntry) isExpired(now time.Time) bool {
	return uint32(now.Unix()) >= e.expiry
}

func (e *sessionEntry) isValid(now time.Time) bool {
	return !e.hasWaiter() && !e.isExpired(now)
}

// deviceData holds sessions entries and gc metadata associated to a single
// device id.
type deviceData struct {
	// Entries sorted by expiry timestamp in ascending order (e.g. new sessions
	// are appended).
	entries []sessionEntry

	// Associated gc job.
	gcData gcJob
}

type service struct {
	logger  zerolog.Logger
	cfg     Config
	metrics metrics
	mu      sync.Mutex
	devices map[uint64]*deviceData

	// GC priority queue.
	gcQueue gcQueue

	// Internal clock.
	now atomic.Pointer[time.Time]
}

// ProvideService is a wire provider for in memory session storage.
func ProvideService(
	logger zerolog.Logger,
	cfg Config,
	promRegistry *prometheus.Registry,
) Service {
	logger = logger.With().
		Str("service", "sessionstorage").
		Dur("gc_interval", cfg.gcInterval).
		Uint64("max_sessions_per_visitor", cfg.maxSessionsPerVisitor).
		Int("device_expiry_percentile", cfg.deviceExpiryPercentile).
		Dur("session_inactive_ttl", cfg.sessionInactiveTtl).
		Logger()

	service := &service{
		logger:  logger,
		cfg:     cfg,
		metrics: newMetrics(promRegistry),
		mu:      sync.Mutex{},
		devices: make(map[uint64]*deviceData),
		gcQueue: gcQueue{},
	}
	now := time.Now()
	service.now.Store(&now)
	heap.Init(&service.gcQueue)

	go service.gcLoop()

	logger.Info().Msg("in memory session storage configured")

	return service
}

// findDevicePExpiry finds device configured percentile expiry and update
// associated gc job.
func (s *service) updateDevicePExpiry(device *deviceData) {
	// Find percentile expiry.
	i := (len(device.entries) - 1) * s.cfg.deviceExpiryPercentile / 100

	// Update percentile expiry.
	device.gcData.pExpiry = device.entries[i].expiry
	heap.Fix(&s.gcQueue, device.gcData.jobIndex)
}

// getSession retrieves a pointer to sessionData associated to given device id
// and latest path.
// Note that this function can return a pointer to a session entry without
// session, or an expired one.
// If you want to retrieve a valid session, use getValidSessionEntry instead.
// You must hold mutex while calling this function.
func (s *service) getSession(deviceId uint64, latestPath string) (*sessionEntry, *deviceData, int) {
	device := s.devices[deviceId]
	if device == nil {
		return nil, nil, -1
	}

	for i, entry := range device.entries {
		if entry.latestUri.Path() == latestPath {
			return &device.entries[i], device, i
		}
	}

	return nil, nil, -1
}

// getValidSessionEntry retrieves a pointer to a valid session entry. This
// function updates returned session expiry.
// You must hold mutex while calling this function.
func (s *service) getValidSessionEntry(deviceId uint64, latestPath string) *sessionEntry {
	assert.Locked(&s.mu)
	session, device, i := s.getSession(deviceId, latestPath)
	if session == nil || !session.isValid(*s.now.Load()) {
		return nil
	}

	return s.updateDeviceSessionExpiry(device, *session, i)
}

// insertNewDevice inserts a new device within sessionstorage, overriding existing
// entry if any.
// You must holds mutex while calling this function.
func (s *service) insertNewDevice(deviceId uint64, firstSession sessionEntry) {
	assert.Locked(&s.mu)

	data := &deviceData{
		entries: []sessionEntry{firstSession},
		gcData: gcJob{
			deviceId: deviceId,
			pExpiry:  firstSession.expiry,
		},
	}

	s.devices[deviceId] = data
	heap.Push(&s.gcQueue, &data.gcData)
}

// insertDeviceSession adds a session to an existing device without checking if
// device reached its limit.
// You must holds mutex while calling this function.
func (s *service) insertDeviceSession(device *deviceData, newSession sessionEntry) {
	assert.Locked(&s.mu)

	// newSession.expiry is always greater that latest entry except for waiter
	// session (see WaitSession).
	if newSession.expiry >= device.entries[len(device.entries)-1].expiry {
		device.entries = append(device.entries, newSession)
	} else {
		i, _ := slices.BinarySearchFunc(device.entries, newSession.expiry, func(session sessionEntry, targetExpiry uint32) int {
			if session.expiry == targetExpiry {
				return 0
			} else if session.expiry < targetExpiry {
				return -1
			} else {
				return 1
			}
		})
		device.entries = slices.Insert(device.entries, i, newSession)
	}

	s.updateDevicePExpiry(device)
}

// updateDeviceSessionExpiry updates expiry of the given session and returns
// a new pointer to it. This function invalidates old pointer to session.
// You must holds mutex while calling this function.
func (s *service) updateDeviceSessionExpiry(device *deviceData, session sessionEntry, sessionIndex int) *sessionEntry {
	assert.Locked(&s.mu)

	session.expiry = s.newExpiry()

	// Move session to end of slice to keep expiry ordering.
	device.entries = slices.Delete(device.entries, sessionIndex, sessionIndex+1)
	index := len(device.entries)
	device.entries = append(device.entries, session)

	s.updateDevicePExpiry(device)

	return &device.entries[index]
}

// InsertSession implements Service.
func (s *service) InsertSession(deviceId uint64, session event.Session) bool {
	var waiterEntry *sessionEntry

	s.mu.Lock()
	newEntry := sessionEntry{
		Session:   session,
		latestUri: session.PageUri,
		wait:      nil,
		expiry:    s.newExpiry(),
	}

	deviceData, deviceFound := s.devices[deviceId]
	if !deviceFound {
		// New device, first session.

		s.insertNewDevice(deviceId, newEntry)
	} else if len(deviceData.entries) >= int(s.cfg.maxSessionsPerVisitor) {
		// Maximal number of session per visitor/device reached.

		s.mu.Unlock()
		// Prevent visitor from creating too many sessions.
		return false
	} else {
		// Device exists, insert the new session.

		// Check if someone is waiting on this session.
		for i, entry := range deviceData.entries {
			if entry.hasWaiter() && entry.latestUri.Path() == newEntry.latestUri.Path() {
				close(entry.wait) // Notify waiter.
				entry.wait = nil
				waiterEntry = &deviceData.entries[i]
				break
			}
		}

		// Update entry if session had waiter.
		if waiterEntry != nil {
			*waiterEntry = newEntry
		} else {
			s.insertDeviceSession(deviceData, newEntry)
		}
	}
	s.mu.Unlock()

	// Compute metrics.
	if waiterEntry == nil {
		s.metrics.sessionsCounter.With(prometheus.Labels{"type": "inserted"}).Inc()
	}
	if !deviceFound {
		s.metrics.devicesCounter.With(prometheus.Labels{"type": "inserted"}).Inc()
	}

	return true
}

// AddPageview implements Service.
func (s *service) AddPageview(deviceId uint64, referrer event.ReferrerUri, uri uri.Uri) (event.Session, bool) {
	s.mu.Lock()
	entry := s.getValidSessionEntry(deviceId, referrer.Path())
	if entry == nil {
		s.mu.Unlock()
		return event.Session{}, false
	}

	entry.latestUri = uri
	entry.Session.PageviewCount++

	// Copy before releasing lock.
	sess := entry.Session
	s.mu.Unlock()

	return sess, true
}

// IdentifySession implements Service.
func (s *service) IdentifySession(deviceId uint64, pageUri uri.Uri, visitorId string) (event.Session, bool) {
	s.mu.Lock()
	entry := s.getValidSessionEntry(deviceId, pageUri.Path())
	if entry == nil {
		s.mu.Unlock()
		return event.Session{}, false
	}

	// Update visitor id.
	entry.Session.VisitorId = visitorId

	// Copy before releasing lock.
	sess := entry.Session
	s.mu.Unlock()

	return sess, true
}

// WaitSession implements Service.
func (s *service) WaitSession(deviceId uint64, pageUri uri.Uri, timeout time.Duration) (event.Session, bool) {
	s.mu.Lock()
	currentSession, deviceData, sessionIndex := s.getSession(deviceId, pageUri.Path())

	// Valid session.
	if currentSession != nil && currentSession.isValid(*s.now.Load()) {
		s.mu.Unlock()
		return currentSession.Session, true
	} else if timeout == time.Duration(0) { // Entry not found and timeout is 0s.
		s.mu.Unlock()
		return event.Session{}, false
	}

	var wait <-chan struct{}

	// Create entry with a wait channel.
	if currentSession == nil {
		newSession := sessionEntry{
			Session:   event.Session{},
			latestUri: pageUri,
			wait:      make(chan struct{}),
			expiry:    uint32(time.Now().Add(timeout).Unix()),
		}
		deviceData, deviceFound := s.devices[deviceId]
		if !deviceFound {
			s.insertNewDevice(deviceId, newSession)
		} else {
			s.insertDeviceSession(deviceData, newSession)
			deviceData.entries = append(deviceData.entries, newSession)
		}

		wait = newSession.wait
		s.mu.Unlock()
		s.metrics.sessionsCounter.With(prometheus.Labels{"type": "inserted"}).Inc()
		if !deviceFound {
			s.metrics.devicesCounter.With(prometheus.Labels{"type": "inserted"}).Inc()
		}
	} else if currentSession.hasWaiter() { // Entry exists with wait channel.
		currentSession = s.updateDeviceSessionExpiry(deviceData, *currentSession, sessionIndex)
		wait = currentSession.wait
		s.mu.Unlock()
	}

	s.metrics.sessionsWait.Inc()
	defer s.metrics.sessionsWait.Dec()

	deadlineCh := time.After(timeout)
	select {
	case <-wait:
	case <-deadlineCh:
		return event.Session{}, false
	}

	// Retrieve session.
	s.mu.Lock()
	entry := s.getValidSessionEntry(deviceId, pageUri.Path())
	s.mu.Unlock()

	// Session may have expired.
	if entry != nil {
		return entry.Session, true
	}

	return event.Session{}, false
}

// session garbage collector loop.
func (s *service) gcLoop() {
	tick := time.NewTicker(s.cfg.gcInterval)

	for {
		now := <-tick.C

		s.now.Store(&now)
		s.metrics.gcCycle.Inc()

		// Wait until there is job in gcQueue.
		s.mu.Lock()
		if len(s.gcQueue) == 0 {
			s.mu.Unlock()
			continue
		}

		nowTs := uint32(now.Unix())

		// Peek job.
		job := s.gcQueue[0]

		// Job hasn't expired yet.
		if job.pExpiry > nowTs {
			s.mu.Unlock()
			continue
		}

		// Job has expired, collect garbage.

		device := s.devices[job.deviceId]
		expiredSessions := 0
		var expiredSessionsPageviewCounts []uint16
		for i, session := range device.entries {
			if session.expiry > nowTs {
				break
			}
			expiredSessionsPageviewCounts = append(expiredSessionsPageviewCounts, session.Session.PageviewCount)
			expiredSessions = i + 1
		}

		// Delete device if all associated sessions are expired.
		deleteDevice := expiredSessions == len(device.entries)

		if deleteDevice {
			// Remove device and associated gc job.
			delete(s.devices, job.deviceId)
			heap.Remove(&s.gcQueue, job.jobIndex)
		} else if expiredSessions > 0 {
			device.entries = slices.Delete(device.entries, 0, expiredSessions)
			s.updateDevicePExpiry(device)
		}
		s.mu.Unlock()

		// Update metrics.
		if deleteDevice {
			s.metrics.devicesCounter.
				With(prometheus.Labels{"type": "deleted"}).
				Inc()
		}
		s.metrics.sessionsCounter.
			With(prometheus.Labels{"type": "expired"}).
			Add(float64(expiredSessions))
		for _, pvCount := range expiredSessionsPageviewCounts {
			s.metrics.sessionsPageviews.Observe(float64(pvCount))
		}
	}
}

func (s *service) newExpiry() uint32 {
	return uint32(s.now.Load().Add(s.cfg.sessionInactiveTtl).Unix())
}
