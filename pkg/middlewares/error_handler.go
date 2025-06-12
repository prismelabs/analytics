package middlewares

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type ErrorHandler fiber.Handler

// ProvideErrorHandler is a wire provider for a simple error handler middleware.
func ProvideErrorHandler(promRegistry *prometheus.Registry, logger zerolog.Logger) ErrorHandler {
	reqsPanics := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_panics_total",
		Help: "Total number of HTTP request that lead to a panic",
	}, []string{"path", "method", "status"})

	// Register metric.
	promRegistry.MustRegister(reqsPanics)

	return func(c *fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				c.Status(fiber.ErrInternalServerError.Code)

				labels := prometheus.Labels{
					"path":   utils.CopyString(c.Route().Path),
					"method": utils.CopyString(c.Method()),
					"status": strconv.Itoa(c.Response().StatusCode()),
				}
				reqsPanics.With(labels).Inc()
				logger.Error().Any("error", err).Msg("http request handler panicked")
			}
		}()

		err := c.Next()

		var fiberErr *fiber.Error
		if err != nil {
			if errors.As(err, &fiberErr) {
				c.Response().SetStatusCode(fiberErr.Code)
			} else if c.Response().StatusCode() == fiber.StatusOK {
				c.Response().SetStatusCode(fiber.StatusInternalServerError)
			}
		}

		return err
	}
}
