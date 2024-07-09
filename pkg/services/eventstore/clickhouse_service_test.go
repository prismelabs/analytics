package eventstore

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestIntegService(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	logger := log.NewLogger("eventstore_service_test", io.Discard, true)
	ch := clickhouse.ProvideCh(logger, config.ClickhouseFromEnv(), clickhouse.ProvideEmbeddedSourceDriver(logger))
	teardownService := teardown.ProvideService()
	cfg := Config{
		MaxBatchSize:      1,
		MaxBatchTimeout:   time.Millisecond,
		RingBuffersFactor: 1,
	}

	t.Run("SinglePageView", func(t *testing.T) {
		promRegistry := prometheus.NewRegistry()
		service := ProvideService(cfg, ch, logger, promRegistry, teardownService)

		// Add event to batch.
		eventTime := time.Now().UTC().Round(time.Second)
		err := service.StorePageView(context.Background(), &event.PageView{
			Timestamp: eventTime,
			PageUri:   testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
			Session: event.Session{
				PageUri:       testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
				ReferrerUri:   event.ReferrerUri{},
				Client:        uaparser.Client{},
				CountryCode:   ipgeolocator.CountryCode{},
				VisitorId:     "singlePageViewTestCase",
				SessionUuid:   uuid.Must(uuid.NewV7()),
				Utm:           event.UtmParams{},
				PageviewCount: 1,
			},
		})
		require.NoError(t, err)

		// Ensure events are stored.
		time.Sleep(10 * time.Millisecond)

		// Ensure event is stored.
		row := ch.QueryRow(
			context.Background(),
			"SELECT timestamp FROM prisme.pageviews WHERE session_uuid IN (SELECT session_uuid FROM prisme.sessions WHERE visitor_id = 'singlePageViewTestCase')",
			eventTime,
		)
		var storedEventTime time.Time
		err = row.Scan(&storedEventTime)
		require.NoError(t, err)
		require.Equal(t, eventTime, storedEventTime)

		// Check metrics.
		labels := prometheus.Labels{"type": "pageview"}
		require.Equal(t, float64(0),
			testutils.CounterValue(t, promRegistry, "eventstore_batch_dropped_total",
				labels))
		require.Equal(t, float64(0),
			testutils.CounterValue(t, promRegistry, "eventstore_batch_retry_total",
				labels))
		require.Equal(t, float64(1),
			testutils.CounterValue(t, promRegistry, "eventstore_events_total",
				labels))
		require.Equal(t, float64(0),
			testutils.CounterValue(t, promRegistry, "eventstore_ring_buffers_dropped_events_total",
				labels))
		require.Greater(t, float64(1),
			testutils.HistogramSumValue(t, promRegistry, "eventstore_send_batch_duration_seconds",
				labels))
		require.Equal(t, float64(1),
			testutils.HistogramSumValue(t, promRegistry, "eventstore_batch_size_events",
				labels))
		require.Equal(t, uint64(1),
			testutils.HistogramBucketValue(t, promRegistry, "eventstore_batch_size_events",
				labels, 10))
	})

	t.Run("MultipleEvents/PageviewsAndCustom", func(t *testing.T) {
		promRegistry := prometheus.NewRegistry()
		service := ProvideService(cfg, ch, logger, promRegistry, teardownService)

		testStartTime := time.Now().UTC()
		// Store events.
		sessionsCount := 10
		for i := 0; i < sessionsCount; i++ {
			sessionUuid := uuid.Must(uuid.NewV7())

			// Pageview to create entry in sessions table.
			eventTime := time.Now().UTC().Round(time.Second)
			err := service.StorePageView(context.Background(), &event.PageView{
				Timestamp: eventTime,
				PageUri:   testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
				Session: event.Session{
					PageUri:       testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
					ReferrerUri:   event.ReferrerUri{},
					Client:        uaparser.Client{},
					CountryCode:   ipgeolocator.CountryCode{},
					VisitorId:     "multipleEventsTestCase",
					SessionUuid:   sessionUuid,
					Utm:           event.UtmParams{},
					PageviewCount: 1,
				},
			})
			require.NoError(t, err)

			// Custom event associated to the same session.
			eventTime = time.Now().UTC().Round(time.Second)
			err = service.StoreCustom(context.Background(), &event.Custom{
				Timestamp: eventTime,
				PageUri:   testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
				Session: event.Session{
					PageUri:       testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
					ReferrerUri:   event.ReferrerUri{},
					Client:        uaparser.Client{},
					CountryCode:   ipgeolocator.CountryCode{},
					VisitorId:     "multipleEventsTestCase",
					SessionUuid:   sessionUuid,
					Utm:           event.UtmParams{},
					PageviewCount: 1,
				},
				Name:   "foo",
				Keys:   []string{},
				Values: []string{},
			})
			require.NoError(t, err)

			// Ensure events are stored.
			time.Sleep(10 * time.Millisecond)
		}

		// Ensure pageviews events are stored.
		{
			row := ch.QueryRow(
				context.Background(),
				"SELECT COUNT(*) FROM prisme.pageviews WHERE timestamp >= $1 AND session_uuid IN (SELECT session_uuid FROM prisme.sessions WHERE visitor_id = 'multipleEventsTestCase')",
				testStartTime,
			)
			var pageviewsCount uint64
			err := row.Scan(&pageviewsCount)
			require.NoError(t, err)
			require.Equal(t, uint64(sessionsCount), pageviewsCount)
		}

		// Ensure custom events are stored.
		{
			row := ch.QueryRow(
				context.Background(),
				"SELECT COUNT(*) FROM prisme.events_custom WHERE timestamp >= $1 AND session_uuid IN (SELECT session_uuid FROM prisme.sessions WHERE visitor_id = 'multipleEventsTestCase')",
				testStartTime,
			)
			var customEventsCount uint64
			err := row.Scan(&customEventsCount)
			require.NoError(t, err)
			require.Equal(t, uint64(sessionsCount), customEventsCount)
		}

		// Check pageview metrics.
		{
			labels := prometheus.Labels{"type": "pageview"}
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "eventstore_batch_dropped_total",
					labels))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "eventstore_batch_retry_total",
					labels))
			require.Equal(t, float64(sessionsCount),
				testutils.CounterValue(t, promRegistry, "eventstore_events_total",
					labels))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "eventstore_ring_buffers_dropped_events_total",
					labels))
			require.Greater(t,
				testutils.HistogramSumValue(t, promRegistry, "eventstore_send_batch_duration_seconds", labels),
				float64(0))
			require.Equal(t, float64(sessionsCount),
				testutils.HistogramSumValue(t, promRegistry, "eventstore_batch_size_events",
					labels))
		}

		// Check custom metrics.
		{
			labels := prometheus.Labels{"type": "custom"}
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "eventstore_batch_dropped_total",
					labels))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "eventstore_batch_retry_total",
					labels))
			require.Equal(t, float64(sessionsCount),
				testutils.CounterValue(t, promRegistry, "eventstore_events_total",
					labels))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "eventstore_ring_buffers_dropped_events_total",
					labels))
			require.Greater(t,
				testutils.HistogramSumValue(t, promRegistry, "eventstore_send_batch_duration_seconds", labels),
				float64(0))
			require.Equal(t, float64(sessionsCount),
				testutils.HistogramSumValue(t, promRegistry, "eventstore_batch_size_events",
					labels))
		}
	})

	t.Run("RingBufferDroppedEvents", func(t *testing.T) {
		promRegistry := prometheus.NewRegistry()
		service := ProvideService(cfg, ch, logger, promRegistry, teardownService)

		// Send hundreds of event without pause.
		for i := 0; i < 100; i++ {
			eventTime := time.Now().UTC().Round(time.Second)
			err := service.StorePageView(context.Background(), &event.PageView{
				Timestamp: eventTime,
				PageUri:   testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
				Session: event.Session{
					PageUri:       testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
					ReferrerUri:   event.ReferrerUri{},
					Client:        uaparser.Client{},
					CountryCode:   ipgeolocator.CountryCode{},
					VisitorId:     "singlePageViewTestCase",
					SessionUuid:   uuid.Must(uuid.NewV7()),
					Utm:           event.UtmParams{},
					PageviewCount: 1,
				},
			})
			require.NoError(t, err)
		}

		// Ensure events are stored.
		time.Sleep(10 * time.Millisecond)

		// Check pageview metrics.
		{
			labels := prometheus.Labels{"type": "pageview"}
			require.Greater(t,
				testutils.CounterValue(t, promRegistry, "eventstore_ring_buffers_dropped_events_total", labels),
				float64(0))
		}
	})
}
