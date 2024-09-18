package sessionstorage

import (
	"sync"
	"time"

	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type AddPageviewResult struct {
	Session           event.Session
	DuplicatePageview bool
}

// Service define an in memory session storage.
type Service interface {
	// InsertSession inserts session in session storage and associate it to the
	// given deviceId.
	InsertSession(deviceId uint64, session event.Session)
	// IncSessionPageviewCount increments pageview and returns it.
	AddPageview(deviceId uint64, pageUri uri.Uri) (AddPageviewResult, bool)
	// IdentifySession updates stored session visitor id. Updated session and
	// boolean found flag are returned.
	IdentifySession(deviceId uint64, visitorId string) (event.Session, bool)
	// WaitForSession retrieves stored session and returns it. If session is not
	// found, it waits until it is created or timeout.
	// Returned boolean flag is false if wait timed out and returned an empty
	// session.
	WaitSession(deviceId uint64, timeout time.Duration) (event.Session, bool)
}

type service struct {
	logger  zerolog.Logger
	cfg     Config
	metrics metrics
	mu      sync.RWMutex
	data    map[uint64]entry
}

type entry struct {
	session   event.Session
	latestUri uri.Uri
	expiry    uint32
	wait      chan struct{}
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
		data:    make(map[uint64]entry),
	}

	go service.gc(cfg.gcInterval)

	logger.Info().Msg("in memory session storage configured")

	return service
}

// getSessionEntry retrieves an entry from the map and returns whether or not
// session exists. An entry may exists but have no session associated (e.g.
// someone is waiting for its creation). This function doesn't check if session
// has expired.
// You must hold mutex while calling this function.
func (s *service) getSessionEntry(deviceId uint64) (entry, bool) {
	entry, ok := s.data[deviceId]
	return entry, ok &&
		entry.wait == nil // Someone is waiting on this session but none exists.
}

// getValidSessionEntry is the same as getSessionEntry but returns true only if
// the session is valid.
// You must hold mutex while calling this function.
func (s *service) getValidSessionEntry(deviceId uint64) (entry, bool) {
	entry, ok := s.getSessionEntry(deviceId)
	return entry, ok && uint32(time.Now().Unix()) < entry.expiry // Not expired session.
}

// InsertSession implements Service.
func (s *service) InsertSession(deviceId uint64, session event.Session) {
	s.mu.Lock()
	currentSession, sessionExists := s.getSessionEntry(deviceId)

	// Store session.
	s.data[deviceId] = entry{
		session:   session,
		latestUri: session.PageUri,
		expiry:    s.newExpiry(),
		wait:      nil,
	}
	s.mu.Unlock()

	// Compute metrics.
	s.metrics.sessionsCounter.With(prometheus.Labels{"type": "inserted"}).Inc()
	if !sessionExists {
		// Notify waiters.
		if currentSession.wait != nil {
			close(currentSession.wait)
		}
		s.metrics.activeSessions.Inc()
	} else {
		s.metrics.sessionsCounter.With(prometheus.Labels{"type": "overwritten"}).Inc()
		s.metrics.sessionsPageviews.Observe(float64(currentSession.session.PageviewCount))
	}
}

// AddPageview implements Service.
func (s *service) AddPageview(deviceId uint64, pageUri uri.Uri) (AddPageviewResult, bool) {
	s.mu.Lock()
	sessionEntry, ok := s.getValidSessionEntry(deviceId)
	// Session not found.
	if !ok {
		s.mu.Unlock()
		return AddPageviewResult{}, false
	}

	// Duplicate pageview.
	if sessionEntry.latestUri.Path() == pageUri.Path() {
		s.mu.Unlock()
		return AddPageviewResult{
			Session:           sessionEntry.session,
			DuplicatePageview: true,
		}, true
	}

	sessionEntry.session.PageviewCount++

	s.data[deviceId] = entry{
		session:   sessionEntry.session,
		latestUri: pageUri,
		expiry:    s.newExpiry(),
	}

	s.mu.Unlock()

	return AddPageviewResult{
		Session: sessionEntry.session,
	}, true
}

// IdentifySession implements Service.
func (s *service) IdentifySession(deviceId uint64, visitorId string) (event.Session, bool) {
	s.mu.Lock()
	sessionEntry, ok := s.getValidSessionEntry(deviceId)
	if !ok {
		s.mu.Unlock()
		return event.Session{}, false
	}

	// No need for update.
	if sessionEntry.session.VisitorId == visitorId {
		s.mu.Unlock()
		return sessionEntry.session, true
	}

	// Update visitor id.
	sessionEntry.session.VisitorId = visitorId
	sessionEntry.expiry = s.newExpiry()
	s.data[deviceId] = sessionEntry
	s.mu.Unlock()

	return sessionEntry.session, true
}

// WaitSession implements Service.
func (s *service) WaitSession(deviceId uint64, timeout time.Duration) (event.Session, bool) {
	s.mu.RLock()
	// We don't use getSessionEntry here as we want to check if entry exists
	// (and not if session exists).
	sessionEntry, ok := s.data[deviceId]
	s.mu.RUnlock()

	// Entry contains a session and hasn't expired.
	if ok && sessionEntry.wait == nil && uint32(time.Now().Unix()) < sessionEntry.expiry {
		return sessionEntry.session, true
	} else if timeout == time.Duration(0) {
		return event.Session{}, false
	}

	// Create entry with a wait channel.
	if !ok {
		sessionEntry.wait = make(chan struct{})
		sessionEntry.expiry = uint32(time.Now().Add(timeout).Unix())

		s.mu.Lock()
		s.data[deviceId] = sessionEntry
		s.mu.Unlock()
	} else if ok && sessionEntry.wait != nil { // Entry exists with wait channel.
		// Update expiry.
		sessionEntry.expiry = uint32(time.Now().Add(timeout).Unix())

		s.mu.Lock()
		s.data[deviceId] = sessionEntry
		s.mu.Unlock()
	}

	// Wait if needed.
	if sessionEntry.wait != nil {
		s.metrics.sessionsWait.Inc()
		defer s.metrics.sessionsWait.Dec()

		deadlineCh := time.After(timeout)
		select {
		case <-sessionEntry.wait:
			break
		case <-deadlineCh:
			return event.Session{}, false
		}
	}

	s.mu.RLock()
	sessionEntry, ok = s.getValidSessionEntry(deviceId)
	s.mu.RUnlock()

	return sessionEntry.session, ok
}

// session garbage collector.
func (s *service) gc(gcInterval time.Duration) {
	ticker := time.NewTicker(gcInterval)
	defer ticker.Stop()
	var expired []uint64
	var expiredSessionPageviews []uint16

	for {
		<-ticker.C

		expired = expired[:0]
		expiredSessionPageviews = expiredSessionPageviews[:0]

		ts := uint32(time.Now().Unix())

		// Collect expired sessions.
		s.mu.RLock()
		for id, v := range s.data {
			if ts >= v.expiry {
				expired = append(expired, id)
			}
		}
		s.mu.RUnlock()

		// Delete expired sessions.
		s.mu.Lock()
		// Double-checked locking.
		// We might have replaced the item in the meantime.
		expiredCounter := 0
		for i := range expired {
			v := s.data[expired[i]]
			if ts >= v.expiry {
				if v.wait != nil {
					expiredCounter++
					expiredSessionPageviews = append(expiredSessionPageviews, v.session.PageviewCount)
				}

				delete(s.data, expired[i])
			}
		}
		s.mu.Unlock()

		// Update metrics.
		s.metrics.sessionsCounter.
			With(prometheus.Labels{"type": "expired"}).
			Add(float64(expiredCounter))
		s.metrics.activeSessions.Sub(float64(expiredCounter))
		for _, pv := range expiredSessionPageviews {
			s.metrics.sessionsPageviews.Observe(float64(pv))
		}
	}
}

func (s *service) newExpiry() uint32 {
	return uint32(time.Now().Add(s.cfg.sessionInactiveTtl).Unix())
}
