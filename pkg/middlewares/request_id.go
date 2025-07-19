package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/prisme"
)

type RequestIdKey struct{}

// RequestId returns request id middleware.
func RequestId(cfg prisme.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestId string

		if cfg.TrustProxy {
			requestId = utils.UnsafeString(c.Request().Header.Peek(cfg.ProxyRequestIdHeader))
		}

		if requestId == "" {
			requestId = utils.UUIDv4()
		}

		c.Locals(RequestIdKey{}, requestId)

		c.Response().Header.Set("X-Prisme-Request-Id", requestId)

		return c.Next()
	}
}
