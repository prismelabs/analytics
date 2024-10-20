package sessionstorage

import (
	"slices"
	"sync"
	"time"

	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// Service define an in memory session storage.
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
	// WaitForSession retrieves stored session and returns it. If session is not
	// found, it waits until it is created or timeout.
	// Returned boolean flag is false if wait timed out and returned an empty
	// session.
	WaitSession(deviceId uint64, pageUri uri.Uri, timeout time.Duration) (event.Session, bool)
}

type entry struct {
	Session   event.Session
	latestUri uri.Uri
	expiry    uint32
	wait      chan struct{}
}

func (e *entry) hasWaiter() bool {
	return e.wait != nil
}

func (e *entry) isExpired() bool {
	return uint32(time.Now().Unix()) >= e.expiry
}

func (e *entry) isValid() bool {
	return !e.hasWaiter() && !e.isExpired()
}

type service struct {
	logger  zerolog.Logger
	cfg     Config
	metrics metrics
	mu      sync.RWMutex
	data    map[uint64][]entry
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
		Dur("session_inactive_ttl", cfg.sessionInactiveTtl).
		Logger()

	service := &service{
		logger:  logger,
		cfg:     cfg,
		metrics: newMetrics(promRegistry),
		mu:      sync.RWMutex{},
		data:    make(map[uint64][]entry),
	}

	go service.gc(cfg.gcInterval)

	logger.Info().Msg("in memory session storage configured")

	return service
}

// getEntry retrieves a pointer to an entry associated to given device id and
// latest path.
// Note that this function can return a pointer to an entry without
// session, or an expired one.
// If you want to retrieve a valid session, use getValidSessionEntry instead.
// You must hold mutex while calling this function.
func (s *service) getEntry(deviceId uint64, latestPath string) *entry {
	entries, ok := s.data[deviceId]
	if !ok {
		return nil
	}

	for i, entry := range entries {
		if entry.latestUri.Path() == latestPath {
			return &entries[i]
		}
	}

	return nil
}

// getValidSessionEntry retrieves a pointer to a valid entry that contains a
// session.
// You must hold mutex while calling this function.
func (s *service) getValidSessionEntry(deviceId uint64, latestPath string) *entry {
	entry := s.getEntry(deviceId, latestPath)
	if entry == nil || !entry.isValid() {
		return nil
	}

	entry.expiry = s.newExpiry()

	return entry
}

// InsertSession implements Service.
func (s *service) InsertSession(deviceId uint64, session event.Session) bool {
	s.mu.Lock()
	newEntry := entry{
		Session:   session,
		latestUri: session.PageUri,
		expiry:    s.newExpiry(),
		wait:      nil,
	}

	sessions, ok := s.data[deviceId]
	// New visitor, first session.
	if !ok {
		sessions = make([]entry, 1)
		sessions[0] = newEntry
		s.data[deviceId] = sessions
	} else if len(sessions) >= int(s.cfg.maxSessionsPerVisitor) {
		s.mu.Unlock()
		// Prevent visitor from creating too many sessions.
		return false
	} else {
		// New session only.

		var waiterEntry *entry
		// Check if someone is waiting on this session.
		for i, sess := range sessions {
			if sess.hasWaiter() && sess.latestUri.Path() == newEntry.latestUri.Path() {
				close(sess.wait) // Notify waiter.
				waiterEntry = &sessions[i]
				break
			}
		}

		// Update entry if session had waiter.
		if waiterEntry != nil {
			*waiterEntry = newEntry
		} else {
			s.data[deviceId] = append(sessions, newEntry)
		}
	}
	s.mu.Unlock()

	// Compute metrics.
	s.metrics.sessionsCounter.With(prometheus.Labels{"type": "inserted"}).Inc()
	s.metrics.activeSessions.Inc()

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
	s.mu.RLock()
	currentEntry := s.getEntry(deviceId, pageUri.Path())
	s.mu.RUnlock()

	// Entry contains a session and hasn't expired.
	if currentEntry != nil && !currentEntry.hasWaiter() && !currentEntry.isExpired() {
		return currentEntry.Session, true
	} else if timeout == time.Duration(0) { // Entry not found and timeout is 0s.
		return event.Session{}, false
	}

	var wait <-chan struct{}

	// Create entry with a wait channel.
	if currentEntry == nil {
		s.mu.Lock()
		newEntry := entry{
			Session:   event.Session{},
			latestUri: pageUri,
			expiry:    uint32(time.Now().Add(timeout).Unix()),
			wait:      make(chan struct{}),
		}
		s.data[deviceId] = append(s.data[deviceId], newEntry)
		wait = newEntry.wait
		s.mu.Unlock()
	} else if currentEntry.hasWaiter() { // Entry exists with wait channel.
		s.mu.Lock()
		currentEntry.expiry = uint32(time.Now().Add(timeout).Unix())
		wait = currentEntry.wait
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
	s.mu.RLock()
	entry := s.getValidSessionEntry(deviceId, pageUri.Path())
	s.mu.RUnlock()

	// Session may have expired.
	if entry != nil {
		return entry.Session, true
	}

	return event.Session{}, false
}

// session garbage collector.
func (s *service) gc(gcInterval time.Duration) {
	type Range struct {
		start, end int
	}
	type ExpiredEntries struct {
		deviceId     uint64
		entriesRange Range
	}

	ticker := time.NewTicker(gcInterval)
	defer ticker.Stop()
	var expiredEntries []ExpiredEntries

	for {
		<-ticker.C

		expiredEntries = expiredEntries[:0]

		ts := uint32(time.Now().Unix())

		// Collect expired sessions.
		s.mu.RLock()

		for id, entries := range s.data {
			var current *ExpiredEntries

			for i, entry := range entries {
				if ts >= entry.expiry {
					if current == nil || current.entriesRange.end != i-1 {
						expiredEntries = append(expiredEntries, ExpiredEntries{id, Range{i, i + 1}})
						current = &expiredEntries[len(expiredEntries)-1]
					} else {
						current.entriesRange.end = i
					}
				}
			}
		}
		s.mu.RUnlock()

		// Delete expired sessions.
		s.mu.Lock()
		// Double-checked locking.
		// We might have replaced the item in the meantime.
		expiredCounter := 0
		for _, expired := range expiredEntries {
			entries := s.data[expired.deviceId]

			for i := expired.entriesRange.start; i < expired.entriesRange.end; i++ {
				if ts >= entries[i].expiry {
					if !entries[i].hasWaiter() {
						expiredCounter++
						s.metrics.sessionsPageviews.Observe(float64(entries[i].Session.PageviewCount))
					}
				} else {
					// Not expired anymore.
					// Split range in two, range start to i (exclusif) is deleted while
					// i to end is appended to expired entries slice.

					// append remaining expired entries.
					if i != expired.entriesRange.end-1 {
						deletedCount := i - expired.entriesRange.start
						expiredEntries = append(expiredEntries, ExpiredEntries{
							deviceId:     expired.deviceId,
							entriesRange: Range{i + 1 - deletedCount, expired.entriesRange.end - deletedCount},
						})
					}

					// Remove entries.
					s.data[expired.deviceId] = slices.Delete(entries, expired.entriesRange.start, i)
					break
				}
			}

			// If entries slice is empty, remove it from map.
			if len(s.data[expired.deviceId]) == 0 {
				delete(s.data, expired.deviceId)
			}
		}
		s.mu.Unlock()

		// Update metrics.
		s.metrics.sessionsCounter.
			With(prometheus.Labels{"type": "expired"}).
			Add(float64(expiredCounter))
		s.metrics.activeSessions.Sub(float64(expiredCounter))
	}
}

func (s *service) newExpiry() uint32 {
	return uint32(time.Now().Add(s.cfg.sessionInactiveTtl).Unix())
}
