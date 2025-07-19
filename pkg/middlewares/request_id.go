package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/prisme"
)

type RequestId fiber.Handler

type RequestIdKey struct{}

// ProvideRequestId define a wire provider for request id middleware.
func ProvideRequestId(cfg prisme.Config) RequestId {
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
