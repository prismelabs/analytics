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
	// Returned boolean flag is false is no session was found.
	GetSession(visitorId string) (event.Session, bool)
	// UpsertSession updates or insert a new session and return true if upsert
	// wasn't ignored. Upsert are ignored if more recent session already exists.
	// A session is considered more recent if stored session UUID differ and
	// session.SessionTime() is more recent or if sessions shares same session
	// UUIDs but session.Version() is greater than stored one.
	UpsertSession(session event.Session) bool
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
func (s *service) GetSession(visitorId string) (event.Session, bool) {
	s.mu.RLock()
	entry, ok := s.data[visitorId]
	s.mu.RUnlock()

	// Expired session.
	if ok && uint32(time.Now().Unix()) >= entry.expiry {
		s.metrics.getSessionsMiss.Inc()
		return event.Session{}, false
	} else if !ok {
		s.metrics.getSessionsMiss.Inc()
	}

	return entry.session, ok
}

// UpsertSession implements Service.
func (s *service) UpsertSession(session event.Session) bool {
	s.mu.Lock()
	sessionEntry, sessionExists := s.data[session.VisitorId]

	sameSession := session.SessionUuid == sessionEntry.session.SessionUuid
	newSession := !sessionExists || (!sameSession && session.SessionTime().Sub(sessionEntry.session.SessionTime()) > 0)
	updatedSession := sessionExists && sameSession && session.Version() > sessionEntry.session.Version()

	upsert := newSession || updatedSession
	if upsert {
		s.data[session.VisitorId] = entry{
			session: session,
			expiry:  uint32(time.Now().Add(s.cfg.sessionInactiveTtl).Unix()),
		}
	}
	s.mu.Unlock()

	// New session.
	if newSession {
		s.metrics.sessionsCounter.With(prometheus.Labels{"type": "inserted"}).Inc()
	}

	// New session but no overwrite.
	if newSession && !sessionExists {
		s.metrics.activeSessions.Inc()
	}

	// New session with overwrite.
	if newSession && sessionExists {
		s.metrics.sessionsPageviews.Observe(float64(sessionEntry.session.Pageviews))
	}

	return upsert
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
				expiredSessionPageviews = append(expiredSessionPageviews, v.session.Pageviews)

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
