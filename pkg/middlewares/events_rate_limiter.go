package middlewares

import (
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/storage"
	"github.com/prismelabs/analytics/pkg/prisme"
)

// EventsRateLimiter returns a rate limiter middleware for /api/*/events/* handlers.
func EventsRateLimiter(cfg prisme.Config, storage storage.Storage) fiber.Handler {
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
