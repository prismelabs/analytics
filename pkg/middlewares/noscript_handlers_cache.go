package middlewares

import "github.com/gofiber/fiber/v2"

type NoscriptHandlersCache fiber.Handler

// ProvideNoscriptHandlersCache is a wire provider for caching middleware used
// for GET /api/v1/noscript/events/... handlers.
func ProvideNoscriptHandlersCache() NoscriptHandlersCache {
	return func(c *fiber.Ctx) error {
		c.Response().Header.Add(fiber.HeaderCacheControl, "no-store, no-cache, max-age=0, must-revalidate, proxy-revalidate")
		return c.Next()
	}
}
