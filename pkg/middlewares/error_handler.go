package middlewares

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
)

// ErrorHandler returns a simple error handler middleware.
func ErrorHandler(promRegistry *prometheus.Registry, logger log.Logger) fiber.Handler {
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
				logger.Error("http request handler panicked", "error", err)
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
