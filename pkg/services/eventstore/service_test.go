//go:build !race

package eventstore

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/clickhouse"
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

func TestIntegNoRaceDetectorService(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	for _, backend := range []string{"chdb", "clickhouse"} {
		t.Run(backend, func(t *testing.T) {
			logger := log.NewLogger("eventstore_service_test", io.Discard, true)
			teardownService := teardown.ProvideService()
			source := clickhouse.ProvideEmbeddedSourceDriver(logger)
			cfg := Config{
				Backend:           backend,
				BackendConfig:     nil,
				MaxBatchSize:      1,
				MaxBatchTimeout:   time.Millisecond,
				RingBuffersFactor: 1,
			}

			t.Run("SinglePageView", func(t *testing.T) {
				promRegistry := prometheus.NewRegistry()
				service := ProvideService(cfg, logger, promRegistry, teardownService, source)

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
				result, err := service.Query(
					context.Background(),
					"SELECT timestamp FROM prisme.pageviews WHERE session_uuid IN (SELECT session_uuid FROM prisme.sessions WHERE visitor_id = 'singlePageViewTestCase')",
					eventTime,
				)
				require.NoError(t, err)
				var storedEventTime time.Time
				err = result.Scan(&storedEventTime)
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

			t.Run("MultipleEvents/Pageviews/Custom/OutboundLinkClick", func(t *testing.T) {
				promRegistry := prometheus.NewRegistry()
				service := ProvideService(cfg, logger, promRegistry, teardownService, source)

				testStartTime := time.Now().UTC()
				// Store events.
				sessionsCount := 10
				for i := 0; i < sessionsCount; i++ {
					sessionUuid := uuid.Must(uuid.NewV7())
					session := event.Session{
						PageUri:       testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
						ReferrerUri:   event.ReferrerUri{},
						Client:        uaparser.Client{},
						CountryCode:   ipgeolocator.CountryCode{},
						VisitorId:     "multipleEventsTestCase",
						SessionUuid:   sessionUuid,
						Utm:           event.UtmParams{},
						PageviewCount: 1,
					}

					// Pageview to create entry in sessions table.
					eventTime := time.Now().UTC().Round(time.Second)
					err := service.StorePageView(context.Background(), &event.PageView{
						Timestamp: eventTime,
						PageUri:   testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
						Session:   session,
					})
					require.NoError(t, err)

					// Custom event associated to the same session.
					eventTime = time.Now().UTC().Round(time.Second)
					err = service.StoreCustom(context.Background(), &event.Custom{
						Timestamp: eventTime,
						PageUri:   testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
						Session:   session,
						Name:      "foo",
						Keys:      []string{},
						Values:    []string{},
					})
					require.NoError(t, err)

					eventTime = time.Now().UTC().Round(time.Second)
					err = service.StoreOutboundLinkClick(context.Background(), &event.OutboundLinkClick{
						Timestamp: eventTime,
						PageUri:   testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
						Session:   session,
						Link:      testutils.Must(uri.Parse)("http://example.com"),
					})
					require.NoError(t, err)

					eventTime = time.Now().UTC().Round(time.Second)
					err = service.StoreFileDownload(context.Background(), &event.FileDownload{
						Timestamp: eventTime,
						PageUri:   testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
						Session:   session,
						FileUrl:   testutils.Must(uri.Parse)("http://mywebsite.localhost/slide.pdf"),
					})
					require.NoError(t, err)

					// Ensure events are stored.
					time.Sleep(50 * time.Millisecond)
				}

				// Ensure pageviews events are stored.
				{
					row, err := service.Query(
						context.Background(),
						"SELECT COUNT(*) FROM prisme.pageviews WHERE timestamp >= $1 AND session_uuid IN (SELECT session_uuid FROM prisme.sessions WHERE visitor_id = 'multipleEventsTestCase')",
						testStartTime,
					)
					require.NoError(t, err)
					var pageviewsCount uint64
					err = row.Scan(&pageviewsCount)
					require.NoError(t, err)
					require.Equal(t, uint64(sessionsCount), pageviewsCount)
				}

				// Ensure custom events are stored.
				{
					row, err := service.Query(
						context.Background(),
						"SELECT COUNT(*) FROM prisme.events_custom WHERE timestamp >= $1 AND session_uuid IN (SELECT session_uuid FROM prisme.sessions WHERE visitor_id = 'multipleEventsTestCase')",
						testStartTime,
					)
					require.NoError(t, err)
					var customEventsCount uint64
					err = row.Scan(&customEventsCount)
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

				// Check outbound link click metrics.
				{
					labels := prometheus.Labels{"type": "outbound_link_click"}
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

				// Check file download metrics.
				{
					labels := prometheus.Labels{"type": "file_download"}
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
				cfg := Config{
					MaxBatchSize:      1_000,
					MaxBatchTimeout:   10 * time.Millisecond,
					RingBuffersFactor: 1,
				}

				service := ProvideService(cfg, logger, promRegistry, teardownService, source)

				// Send hundreds of event without pause.
				for i := 0; i < 10_000; i++ {
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
		})
	}
}
