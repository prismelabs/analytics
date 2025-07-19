package middlewares

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/prisme"
)

type ApiEventsTimeout fiber.Handler

// ProvideApiEventsTimeout is a wire provider for timeout middleware used on
// /api/v1/events routes.
func ProvideApiEventsTimeout(cfg prisme.Config) ApiEventsTimeout {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.ApiEventsTimeout)
		c.SetUserContext(ctx)

		err := c.Next()

		cancel()

		return err
	}
}
