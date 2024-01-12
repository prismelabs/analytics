package middlewares

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
	"github.com/rs/zerolog"
)

type AccessLog fiber.Handler

// ProvideAccessLog define a wire provider for AccessLog middleware.
func ProvideAccessLog(cfg config.Server, logger log.Logger) AccessLog {
	// Open access log file.
	accessLogFile, err := os.OpenFile(cfg.AccessLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		logger.Panic().Err(err).Msgf("failed to open access log file: %v", cfg.AccessLog)
	}

	accessLogger := log.NewLogger("access_log", accessLogFile, cfg.Debug)
	log.TestLoggers(accessLogger)

	return accessLog(accessLogger)
}

func accessLog(logger log.Logger) fiber.Handler {
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
