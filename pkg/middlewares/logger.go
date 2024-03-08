package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type Logger fiber.Handler

type LoggerKey struct{}

// ProvideLogger define a wire provider for logger middleware.
func ProvideLogger(logger zerolog.Logger) Logger {
	return func(c *fiber.Ctx) error {
		requestId := c.Locals(RequestIdKey{}).(string)
		logger = logger.With().Str("request_id", requestId).Logger()

		c.Locals(LoggerKey{}, logger)
		return c.Next()
	}
}
