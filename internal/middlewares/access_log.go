package middlewares

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prismelabs/prismeanalytics/internal/log"
	"github.com/rs/zerolog"
)

func AccessLog(logger log.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			statusCode := c.Response().Status
			level := zerolog.InfoLevel
			if err != nil {
				level = zerolog.ErrorLevel
			}

			logger.WithLevel(level).
				Str("request_id", c.Get(RequestIdKey).(string)).
				Dur("duration_ms", time.Since(start)).
				Str("source_ip", c.RealIP()).
				Str("method", c.Request().Method).
				Str("path", c.Request().URL.Path).
				Int("status_code", statusCode).
				Err(err).
				Msg("request handled")

			return err
		}
	}
}
