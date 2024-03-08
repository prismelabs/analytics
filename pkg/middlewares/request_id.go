package middlewares

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/config"
)

type RequestId fiber.Handler

type RequestIdKey struct{}

// ProvideRequestId define a wire provider for request id middleware.
func ProvideRequestId(cfg config.Server) RequestId {
	return func(c *fiber.Ctx) error {
		var requestId string

		if cfg.TrustProxy {
			requestId = utils.UnsafeString(c.Request().Header.Peek("X-Request-Id"))
		}

		if requestId == "" {
			requestId = utils.UUIDv4()
		}

		c.Locals(RequestIdKey{}, requestId)
		c.SetUserContext(context.WithValue(c.UserContext(), RequestIdKey{}, requestId))

		return c.Next()
	}
}
