package eventstore

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prismelabs/analytics/pkg/testutils"
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

	t.Run("SinglePageView", func(t *testing.T) {
		promRegistry := prometheus.NewRegistry()
		service := ProvideService(Config{
			MaxBatchSize:    1,
			MaxBatchTimeout: time.Millisecond,
		}, ch, logger, promRegistry, teardownService)

		// Add event to batch.
		eventTime := time.Now().UTC().Round(time.Second)
		err := service.StorePageView(context.Background(), &event.PageView{
			Timestamp:   eventTime,
			PageUri:     event.Uri{},
			ReferrerUri: event.ReferrerUri{},
			Client:      uaparser.Client{},
			CountryCode: ipgeolocator.CountryCode{},
			VisitorId:   "singlePageViewTestCase",
		})
		require.NoError(t, err)

		// Wait for event to be stored.
		time.Sleep(10 * time.Millisecond)

		// Ensure event is stored.
		row := ch.QueryRow(
			context.Background(),
			"SELECT timestamp FROM prisme.events_pageviews WHERE timestamp = $1 AND visitor_id = 'singlePageViewTestCase'",
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

	t.Run("MultipleEvents", func(t *testing.T) {
		promRegistry := prometheus.NewRegistry()
		service := ProvideService(Config{
			MaxBatchSize:    10,
			MaxBatchTimeout: time.Millisecond,
		}, ch, logger, promRegistry, teardownService)

		testStartTime := time.Now().UTC()
		// Store events.
		totalEventsCount := 10
		for i := 0; i < totalEventsCount; i++ {
			if i%2 == 0 {
				// Add event to batch.
				eventTime := time.Now().UTC().Round(time.Second)
				err := service.StoreCustom(context.Background(), &event.Custom{
					Timestamp:   eventTime,
					PageUri:     event.Uri{},
					ReferrerUri: event.ReferrerUri{},
					Client:      uaparser.Client{},
					CountryCode: ipgeolocator.CountryCode{},
					VisitorId:   "multipleEventsTestCase",
				})
				require.NoError(t, err)
			} else {
				// Add event to batch.
				eventTime := time.Now().UTC().Round(time.Second)
				err := service.StorePageView(context.Background(), &event.PageView{
					Timestamp:   eventTime,
					PageUri:     event.Uri{},
					ReferrerUri: event.ReferrerUri{},
					Client:      uaparser.Client{},
					CountryCode: ipgeolocator.CountryCode{},
					VisitorId:   "multipleEventsTestCase",
				})
				require.NoError(t, err)
			}

			// Wait for batch time out
			time.Sleep(2 * time.Millisecond)
		}

		// Ensure pageviews events are stored.
		{
			row := ch.QueryRow(
				context.Background(),
				"SELECT COUNT(*) FROM prisme.events_pageviews WHERE timestamp >= $1 AND visitor_id = 'multipleEventsTestCase'",
				testStartTime,
			)
			var pageviewsCount uint64
			err := row.Scan(&pageviewsCount)
			require.NoError(t, err)
			require.Equal(t, uint64(totalEventsCount/2), pageviewsCount)
		}

		// Ensure custom events are stored.
		{
			row := ch.QueryRow(
				context.Background(),
				"SELECT COUNT(*) FROM prisme.events_custom WHERE timestamp >= $1 AND visitor_id = 'multipleEventsTestCase'",
				testStartTime,
			)
			var customEventsCount uint64
			err := row.Scan(&customEventsCount)
			require.NoError(t, err)
			require.Equal(t, uint64(totalEventsCount/2), customEventsCount)
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
			require.Equal(t, float64(totalEventsCount/2),
				testutils.CounterValue(t, promRegistry, "eventstore_events_total",
					labels))
			require.Greater(t, float64(1),
				testutils.HistogramSumValue(t, promRegistry, "eventstore_send_batch_duration_seconds",
					labels))
			require.Equal(t, float64(totalEventsCount/2),
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
			require.Equal(t, float64(5),
				testutils.CounterValue(t, promRegistry, "eventstore_events_total",
					labels))
			require.Greater(t, float64(1),
				testutils.HistogramSumValue(t, promRegistry, "eventstore_send_batch_duration_seconds",
					labels))
			require.Equal(t, float64(totalEventsCount/2),
				testutils.HistogramSumValue(t, promRegistry, "eventstore_batch_size_events",
					labels))
		}
	})
}
