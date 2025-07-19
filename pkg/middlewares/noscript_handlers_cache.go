package middlewares

import "github.com/gofiber/fiber/v2"

// NoscriptHandlersCache returns a caching middleware used for GET
// /api/v1/noscript/events/... handlers.
func NoscriptHandlersCache() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Response().Header.Add(fiber.HeaderCacheControl, "no-store, no-cache, max-age=0, must-revalidate, proxy-revalidate")
		return c.Next()
	}
}
