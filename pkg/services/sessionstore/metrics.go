package sessionstore

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	gcCycle           prometheus.Counter
	gcDuration        prometheus.Histogram
	devicesCounter    *prometheus.CounterVec
	sessionsWait      prometheus.Gauge
	sessionsCounter   *prometheus.CounterVec
	sessionsPageviews prometheus.Histogram
}

func newMetrics(promRegistry *prometheus.Registry) metrics {
	m := metrics{
		gcCycle: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sessionstore_gc_cycles_total",
			Help: "Number of sessionstore garbage collector cycles",
		}),
		gcDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sessionstore_gc_cycles_duration_ms",
			Help:    "Duration of garbage collector cycles",
			Buckets: []float64{1, 2, 3, 5, 10, 15, 25, 30, 50, 100},
		}),
		devicesCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "sessionstore_devices_total",
			Help: "Number of inserted and deleted devices",
		}, []string{"type"}),
		sessionsWait: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sessionstore_sessions_wait",
			Help: "Number of events waiting for a session",
		}),
		sessionsCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "sessionstore_sessions_total",
			Help: "Number of inserted and expired sessions",
		}, []string{"type"}),
		sessionsPageviews: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sessionstore_sessions_pageviews",
			Help:    "Number of pageviews per sessions",
			Buckets: []float64{1, 2, 3, 5, 10, 15, 25, 30, 50, 100},
		}),
	}

	promRegistry.MustRegister(
		m.gcCycle,
		m.gcDuration,
		m.devicesCounter,
		m.sessionsWait,
		m.sessionsCounter,
		m.sessionsPageviews,
	)

	return m
}
