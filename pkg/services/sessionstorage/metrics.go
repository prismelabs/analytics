package sessionstorage

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	activeSessions    prometheus.Gauge
	sessionsWait      prometheus.Gauge
	sessionsCounter   *prometheus.CounterVec
	sessionsPageviews prometheus.Histogram
	getSessionsMiss   prometheus.Counter
}

func newMetrics(promRegistry *prometheus.Registry) metrics {
	m := metrics{
		activeSessions: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sessionstorage_active_sessions",
			Help: "Active sessions stored in memory",
		}),
		sessionsWait: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "sessionstorage_sessions_wait",
			Help: "Number of events waiting for a session",
		}),
		sessionsCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "sessionstorage_sessions_total",
			Help: "Number of inserted, overwritten and expired sessions",
		}, []string{"type"}),
		sessionsPageviews: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "sessionstorage_sessions_pageviews",
			Help:    "Number of pageviews per sessions",
			Buckets: []float64{1, 2, 3, 5, 10, 15, 25, 30, 50, 100},
		}),
		getSessionsMiss: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sessionstorage_get_session_misses",
			Help: "Number of get sessions call that wasn't found",
		}),
	}

	promRegistry.MustRegister(
		m.activeSessions,
		m.sessionsWait,
		m.sessionsCounter,
		m.sessionsPageviews,
		m.getSessionsMiss,
	)

	return m
}
