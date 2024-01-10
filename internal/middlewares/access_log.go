package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/log"
	"github.com/rs/zerolog"
)

func AccessLog(logger log.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		statusCode := c.Response().StatusCode()
		level := zerolog.InfoLevel
		if err != nil {
			level = zerolog.ErrorLevel
		}

		logger.WithLevel(level).
			Str("request_id", c.Locals(RequestIdKey{}).(string)).
			Dur("duration_ms", time.Since(start)).
			Str("source_ip", c.IP()).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status_code", statusCode).
			Err(err).
			Msg("request handled")

		return err
	}
}

