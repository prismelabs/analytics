package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/rs/zerolog"
)

type Logger fiber.Handler

type LoggerKey struct{}

// ProvideLogger define a wire provider for logger middleware.
func ProvideLogger(logger log.Logger) Logger {
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
