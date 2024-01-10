package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/log"
	"github.com/rs/zerolog"
)

type LoggerKey struct{}

func Logger(logger log.Logger) fiber.Handler {
	appLogger := logger
	return func(c *fiber.Ctx) error {
		logger = appLogger
		logger.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
			requestId := c.Locals(RequestIdKey{}).(string)
			return ctx.Str("request_id", requestId)
		})

		c.Locals(LoggerKey{}, logger)
		err := c.Next()
		return err
	}
}
