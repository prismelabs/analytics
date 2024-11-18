package middlewares

import (
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/storage"
	"github.com/prismelabs/analytics/pkg/config"
)

type EventsRateLimiter fiber.Handler

// ProvideEventsRateLimiter is a wire provider for events endpoints rate limiter.
func ProvideEventsRateLimiter(cfg config.Server, storage storage.Storage) EventsRateLimiter {
	max := 60
	if cfg.Debug {
		max = math.MaxInt
	}

	return limiter.New(limiter.Config{
		Max: max,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		Expiration: time.Minute,
		Storage:    storage,
	})
}
