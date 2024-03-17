package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/prismelabs/analytics/pkg/config"
)

type EventsRateLimiter fiber.Handler

// ProvideEventsRateLimiter is a wire provider for events endpoints rate limiter.
func ProvideEventsRateLimiter(cfg config.Server) EventsRateLimiter {
	return limiter.New(limiter.Config{
		Max: 60,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		Expiration: time.Minute,
	})
}
