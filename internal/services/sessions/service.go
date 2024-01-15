package sessions

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/prismelabs/prismeanalytics/internal/services/users"
)

const (
	UserIdKey = "user_id"
)

var (
	ErrSessionNotFound    = errors.New("session not found")
	errSessionIsAnonymous = errors.New("anonymous session")

	sessionConfig = session.Config{
		Storage:           nil, // in memory storage.
		Expiration:        24 * time.Hour,
		KeyLookup:         "cookie:prisme_session_id",
		CookieDomain:      "",
		CookiePath:        "",
		CookieSecure:      true,
		CookieHTTPOnly:    true,
		CookieSameSite:    "Strict",
		CookieSessionOnly: false,
	}
)

// Service define session management service.
type Service interface {
	CreateSession(*fiber.Ctx, users.UserId) error
	GetSession(*fiber.Ctx) (Session, error)
}

// ProvideService is a wire provider for session service.
func ProvideService() Service {
	store := session.New(sessionConfig)

	return newService(store)
}

func newService(store *session.Store) service {
	return service{store}
}

type service struct {
	store *session.Store
}

// CreateSession implements Service.
func (s service) CreateSession(c *fiber.Ctx, userId users.UserId) error {
	session, err := s.store.Get(c)
	if err != nil {
		return fmt.Errorf("failed to retrieve/create session: %w", err)
	}

	// User already have a session, resets it first.
	if !session.Fresh() {
		session.Reset()
	}

	session.Set(UserIdKey, userId.String())

	err = session.Save()
	if err != nil {
		return fmt.Errorf("failed to save session to storage: %w", err)
	}

	return nil
}

// GetSession implements Service.
func (s service) GetSession(c *fiber.Ctx) (Session, error) {
	session, err := s.store.Get(c)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve/create session: %w", err)
	}

	if session.Fresh() {
		return nil, ErrSessionNotFound
	}

	userSess, err := userSessionFromSession(session)
	if err != nil {
		if errors.Is(err, errAnonymousSession) {
			// Destroy anonymous session.
			_ = session.Destroy()
		}
		return nil, fmt.Errorf("failed to create user session from fiber session: %w", err)
	}

	return userSess, nil
}
