package middlewares

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/options"
)

// AccessLog returns an access log middleware.
func AccessLog(cfg options.Server, logger log.Logger) fiber.Handler {
	// Create access logger.
	var accessLogWriter io.Writer
	switch cfg.AccessLog {
	case "/dev/stdout":
		accessLogWriter = os.Stdout
	case "/dev/stderr":
		accessLogWriter = os.Stderr
	default:
		f, err := os.OpenFile(
			cfg.AccessLog,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			os.ModePerm,
		)
		if err != nil {
			logger.Fatal("failed to open access log file", err, "file", cfg.AccessLog)
		}
		accessLogWriter = f
	}
	accessLogger := log.New("access_log", accessLogWriter, false)
	err := accessLogger.TestOutput()
	if err != nil {
		logger.Fatal("failed to write to access log file", err)
	}

	return accessLog(accessLogger)
}

func accessLog(accessLogger log.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		statusCode := c.Response().StatusCode()
		level := log.LevelInfo
		if err != nil {
			level = log.LevelError
		}

		accessLogger.Log(context.Background(),
			level,
			"",
			"request_id", c.Locals(RequestIdKey{}).(string),
			"duration_ms", time.Since(start).Milliseconds(),
			"source_ip", c.IP(),
			"method", c.Method(),
			"path", c.Path(),
			"status_code", statusCode,
			"error", err,
		)

		// We handled error, no need to return it.
		return nil
	}
}
