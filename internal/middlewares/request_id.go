package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/prismeanalytics/internal/config"
)

type RequestIdKey struct{}

func RequestId(cfg config.Server) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestId string

		if cfg.TrustProxy {
			requestId = utils.UnsafeString(c.Request().Header.Peek("X-Request-Id"))
		}

		if requestId == "" {
			requestId = utils.UUIDv4()
		}

		c.Locals(RequestIdKey{}, requestId)

		return c.Next()
	}
}

