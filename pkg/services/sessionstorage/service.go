package sessionstorage

import (
	"sync"
	"time"

	"github.com/prismelabs/analytics/pkg/event"
	"github.com/rs/zerolog"
)

// Service define an in memory session storage.
type Service interface {
	// GetSession retrieves stored session and returns it.
	// Returned boolean flag is false is no session was found.
	GetSession(visitorId string) (event.Session, bool)
	// UpsertSession updates or insert a new session.
	// UpsertSession call is ignored and return false if stored session has
	// higher version number.
	UpsertSession(event.Session) bool
	// DeleteSession deletes a session.
	DeleteSession(visitorId string)
}

type service struct {
	logger zerolog.Logger
	cfg    Config
	mu     sync.RWMutex
	data   map[string]entry
}

type entry struct {
	session event.Session
	expiry  uint32
}

// ProvideService is a wire provider for in memory session storage.
func ProvideService(logger zerolog.Logger, cfg Config) Service {
	logger = logger.With().
		Str("service", "sessionstorage").
		Dur("gc_interval", cfg.gcInterval).
		Dur("session_inactive_ttl", cfg.sessionInactiveTtl).
		Logger()

	service := &service{
		logger: logger,
		cfg:    cfg,
		mu:     sync.RWMutex{},
		data:   map[string]entry{},
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
		return event.Session{}, false
	}

	return entry.session, ok
}

// UpsertSession implements Service.
func (s *service) UpsertSession(session event.Session) bool {
	upserted := false

	s.mu.Lock()
	sessionEntry, ok := s.data[session.VisitorId]
	if !ok || sessionEntry.session.Version() < session.Version() {
		s.data[session.VisitorId] = entry{
			session: session,
			expiry:  uint32(time.Now().Add(s.cfg.sessionInactiveTtl).Unix()),
		}
		upserted = true
	}
	s.mu.Unlock()

	return upserted
}

// DeleteSession implements Service.
func (s *service) DeleteSession(visitorId string) {
	s.mu.Lock()
	delete(s.data, visitorId)
	s.mu.Unlock()
}

// session garbage collector.
func (s *service) gc(gcInterval time.Duration) {
	ticker := time.NewTicker(gcInterval)
	defer ticker.Stop()
	var expired []string

	for {
		select {
		case <-ticker.C:
			ts := uint32(time.Now().Unix())
			expired = expired[:0]
			s.mu.RLock()
			for id, v := range s.data {
				if v.expiry != 0 && ts >= v.expiry {
					expired = append(expired, id)
				}
			}
			s.mu.RUnlock()
			s.mu.Lock()
			// Double-checked locking.
			// We might have replaced the item in the meantime.
			for i := range expired {
				v := s.data[expired[i]]
				if v.expiry != 0 && ts >= v.expiry {
					delete(s.data, expired[i])
				}
			}
			s.mu.Unlock()
		}
	}
}
