package middlewares

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/options"
)

// ApiEventsTimeout returns a timeout middleware used on /api/*/events/*
// handlers.
func ApiEventsTimeout(cfg options.Server) fiber.Handler {
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
