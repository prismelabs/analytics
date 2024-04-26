package middlewares

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics fiber.Handler

// ProvideMetrics is a wire provider for HTTP metrics middleware.
func ProvideMetrics(promRegistry *prometheus.Registry) Metrics {
	activeReqs := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "http_active_requests",
		Help:        "Active HTTP requests",
		ConstLabels: map[string]string{},
	})

	reqsProcessed := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP request processed",
	}, []string{"path", "method", "status"})

	reqsDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_requests_duration_seconds",
		Help:    "HTTP requests duration histogram",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.3, 0.4, 0.5, 1, 5, 10},
	}, []string{"path", "method", "status"})

	// Register metrics.
	promRegistry.MustRegister(activeReqs, reqsProcessed, reqsDuration)

	return func(c *fiber.Ctx) error {
		activeReqs.Inc()
		// Defer so active reqs is still decremented in case of panic.
		defer activeReqs.Dec()

		start := time.Now()

		// Process request.
		err := c.Next()

		labels := prometheus.Labels{
			"path":   c.Route().Path,
			"method": c.Method(),
			"status": strconv.Itoa(c.Response().StatusCode()),
		}

		reqsProcessed.With(labels).Add(1)
		reqsDuration.With(labels).Observe(time.Since(start).Seconds())

		return err
	}
}
