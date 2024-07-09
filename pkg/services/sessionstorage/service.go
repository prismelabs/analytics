package sessionstorage

import (
	"sync"
	"time"

	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// Service define an in memory session storage.
type Service interface {
	// GetSession retrieves stored session and returns it.
	// Returned boolean flag is false if no session was found.
	GetSession(deviceId string) (event.Session, bool)
	// InsertSession inserts session in session storage and associate it to the
	// given deviceId.
	InsertSession(deviceId string, session event.Session)
	// IncSessionPageviewCount increments pageview and returns it.
	IncSessionPageviewCount(deviceId string) (event.Session, bool)
	// IdentifySession updates stored session visitor id. Updated session and
	// boolean found flag are returned.
	IdentifySession(deviceId string, visitorId string) (event.Session, bool)
	// WaitForSession retrieves stored session and returns it. If session is not
	// found, it waits until it is created or timeout.
	// Returned boolean flag is false if wait timed out and returned an empty
	// session.
	WaitSession(deviceId string, timeout time.Duration) (event.Session, bool)
}

type service struct {
	logger  zerolog.Logger
	cfg     Config
	metrics metrics
	mu      sync.RWMutex
	data    map[string]entry
}

type entry struct {
	session event.Session
	expiry  uint32
	wait    chan struct{}
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
		data:    map[string]entry{},
	}

	go service.gc(cfg.gcInterval)

	logger.Info().Msg("in memory session storage configured")

	return service
}

// GetSession implements Service.
func (s *service) GetSession(deviceId string) (event.Session, bool) {
	s.mu.RLock()
	entry, ok := s.data[deviceId]
	s.mu.RUnlock()

	// Expired session.
	if !ok || uint32(time.Now().Unix()) >= entry.expiry {
		s.metrics.getSessionsMiss.Inc()
		return event.Session{}, false
	}

	return entry.session, ok
}

// InsertSession implements Service.
func (s *service) InsertSession(deviceId string, session event.Session) {
	s.mu.Lock()
	currentSession := s.data[deviceId]

	// Store session.
	s.data[deviceId] = entry{
		session: session,
		expiry:  s.newExpiry(),
	}
	s.mu.Unlock()

	// Compute metrics.
	s.metrics.sessionsCounter.With(prometheus.Labels{"type": "inserted"}).Inc()
	if currentSession.expiry == 0 {
		// Notify waiter.
		if currentSession.wait != nil {
			close(currentSession.wait)
		}
		s.metrics.activeSessions.Inc()
	} else {
		s.metrics.sessionsCounter.With(prometheus.Labels{"type": "overwritten"}).Inc()
		s.metrics.sessionsPageviews.Observe(float64(currentSession.session.PageviewCount))
	}
}

// IncSessionPageviewCount implements Service.
func (s *service) IncSessionPageviewCount(deviceId string) (event.Session, bool) {
	s.mu.Lock()
	sessionEntry, ok := s.data[deviceId]
	// Session not found.
	if !ok {
		s.mu.Unlock()
		return event.Session{}, false
	}

	sessionEntry.session.PageviewCount++

	s.data[deviceId] = entry{
		session: sessionEntry.session,
		expiry:  s.newExpiry(),
	}

	s.mu.Unlock()

	return sessionEntry.session, true
}

// IdentifySession implements Service.
func (s *service) IdentifySession(deviceId string, visitorId string) (event.Session, bool) {
	s.mu.Lock()
	sessionEntry, ok := s.data[deviceId]
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
	s.data[deviceId] = entry{
		session: sessionEntry.session,
		expiry:  s.newExpiry(),
	}
	s.mu.Unlock()

	return sessionEntry.session, true
}

// WaitSession implements Service.
func (s *service) WaitSession(deviceId string, timeout time.Duration) (event.Session, bool) {
	s.mu.RLock()
	sessionEntry, entryExists := s.data[deviceId]
	s.mu.RUnlock()

	// Create entry with a wait channel.
	if !entryExists {
		sessionEntry.wait = make(chan struct{})

		s.mu.Lock()
		s.data[deviceId] = sessionEntry
		s.mu.Unlock()
	}

	// Session is active and already exists.
	if entryExists && sessionEntry.expiry >= uint32(time.Now().Unix()) {
		return sessionEntry.session, true
	}

	// Wait if needed.
	if sessionEntry.wait != nil {
		s.metrics.sessionsWait.Inc()
		defer s.metrics.sessionsWait.Dec()

		deadlineCh := time.After(timeout)
		select {
		case <-deadlineCh:
			return event.Session{}, false
		case <-sessionEntry.wait:
		}
	}

	return s.GetSession(deviceId)
}

// session garbage collector.
func (s *service) gc(gcInterval time.Duration) {
	ticker := time.NewTicker(gcInterval)
	defer ticker.Stop()
	var expired []string
	var expiredSessionPageviews []uint16

	for {
		<-ticker.C

		expired = expired[:0]
		expiredSessionPageviews = expiredSessionPageviews[:0]

		ts := uint32(time.Now().Unix())

		// Collect expired sessions.
		s.mu.RLock()
		for id, v := range s.data {
			if v.expiry != 0 && ts >= v.expiry {
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
			if v.expiry != 0 && ts >= v.expiry {
				expiredCounter++
				expiredSessionPageviews = append(expiredSessionPageviews, v.session.PageviewCount)

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
