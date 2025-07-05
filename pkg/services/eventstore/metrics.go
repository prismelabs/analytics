package eventstore

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	batchDropped      prometheus.Counter
	batchRetry        prometheus.Counter
	eventsCounter     prometheus.Counter
	droppedEvents     prometheus.Counter
	sendBatchDuration prometheus.Histogram
	batchSize         prometheus.Histogram
}

func newMetrics(promRegistry *prometheus.Registry) metrics {
	m := metrics{
		batchDropped: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "eventstore_batch_dropped_total",
			Help: "Total number of dropped batch",
		}),
		batchRetry: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "eventstore_batch_retry_total",
			Help: "Total number of retry for send a batch",
		}),
		eventsCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "eventstore_events_total",
			Help: "Number of events sent to ClickHouse",
		}),
		droppedEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "eventstore_ring_buffers_dropped_events_total",
			Help: "Number of events dropped by non blocking ring buffer",
		}),
		sendBatchDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "eventstore_send_batch_duration_seconds",
			Help:    "Duration of send batch operation",
			Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 1, 5, 10, 60, 120},
		}),
		batchSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "eventstore_batch_size_events",
			Help:    "Number of event per batch",
			Buckets: []float64{1, 10, 100, 1_000, 10_000, 25_000, 50_000, 100_000},
		}),
	}

	promRegistry.MustRegister(
		m.batchDropped,
		m.batchRetry,
		m.eventsCounter,
		m.droppedEvents,
		m.sendBatchDuration,
		m.batchSize,
	)

	return m
}
