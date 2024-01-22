package middlewares

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/services/sessions"
)

type WithSession fiber.Handler

type SessionKey struct{}

func ProvideWithSession(sessionsService sessions.Service) WithSession {
	return func(c *fiber.Ctx) error {
		// Ignore api endpoints.
		if strings.HasPrefix(c.Path(), "/api/") {
			return c.Next()
		}

		session, err := sessionsService.GetSession(c)
		if err != nil {
			if errors.Is(err, sessions.ErrSessionNotFound) {
				return c.Redirect("/sign_in")
			}

			return err
		}

		c.Locals(SessionKey{}, session)

		return c.Next()
	}
}
