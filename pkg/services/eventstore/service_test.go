//go:build !race && chdb

package eventstore

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/eventdb"
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

	setup := func(t *testing.T, cfg Config, driver string, driverCfg any) (Service, eventdb.Service, *prometheus.Registry, teardown.Service) {
		logger := log.New("eventstore_service_test", io.Discard, false)
		promRegistry := prometheus.NewRegistry()
		eventDbCfg := eventdb.Config{
			Driver: driver,
		}
		source := clickhouse.EmbeddedSourceDriver(logger)
		teardown := teardown.NewService()
		db, err := eventdb.NewService(eventDbCfg, driverCfg, logger, source, teardown)
		require.NoError(t, err)

		store, err := NewService(cfg, db, logger, promRegistry, teardown)
		require.NoError(t, err)

		return store, db, promRegistry, teardown
	}

	forEachEventDb := func(t *testing.T, cfg Config, fn func(t *testing.T, store Service, db eventdb.Service, promRegistry *prometheus.Registry)) {
		for driver := range eventdb.Drivers() {
			var driverCfg any
			switch driver {
			case "chdb":
				var cfg chdb.Config
				testutils.ConfigueLoad(t, &cfg)
				driverCfg = cfg
			case "clickhouse":
				var cfg clickhouse.Config
				testutils.ConfigueLoad(t, &cfg)
				driverCfg = cfg
			}
			t.Run(driver, func(t *testing.T) {
				store, db, promRegistry, teardown := setup(t, cfg, driver, driverCfg)
				fn(t, store, db, promRegistry)
				require.NoError(t, teardown.Teardown())
			})
		}
	}

	for _, backend := range []string{"clickhouse", "chdb"} {
		t.Run(backend, func(t *testing.T) {
			cfg := Config{
				MaxBatchSize:      1,
				MaxBatchTimeout:   time.Millisecond,
				RingBuffersFactor: 100,
			}

			t.Run("SinglePageView", func(t *testing.T) {
				forEachEventDb(t, cfg, func(t *testing.T, store Service, db eventdb.Service, promRegistry *prometheus.Registry) {
					// Add event to batch.
					eventTime := time.Now().UTC().Round(time.Second)
					err := store.StorePageView(context.Background(), &event.PageView{
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
					time.Sleep(100 * time.Millisecond)

					// Ensure event is stored.
					row := db.QueryRow(
						context.Background(),
						"SELECT timestamp FROM prisme.pageviews WHERE session_uuid IN (SELECT session_uuid FROM prisme.sessions WHERE visitor_id = 'singlePageViewTestCase')",
					)
					var storedEventTime time.Time
					err = row.Scan(&storedEventTime)
					require.NoError(t, err)
					require.Equal(t, eventTime, storedEventTime)

					// Check metrics.
					require.Equal(t, float64(0),
						testutils.CounterValue(t, promRegistry, "eventstore_batch_dropped_total",
							nil))
					require.Equal(t, float64(0),
						testutils.CounterValue(t, promRegistry, "eventstore_batch_retry_total",
							nil))
					require.Equal(t, float64(1),
						testutils.CounterValue(t, promRegistry, "eventstore_events_total",
							nil))
					require.Equal(t, float64(0),
						testutils.CounterValue(t, promRegistry, "eventstore_ring_buffers_dropped_events_total",
							nil))
					require.Greater(t, float64(1),
						testutils.HistogramSumValue(t, promRegistry, "eventstore_send_batch_duration_seconds",
							nil))
					require.Equal(t, float64(1),
						testutils.HistogramSumValue(t, promRegistry, "eventstore_batch_size_events",
							nil))
					require.Equal(t, uint64(1),
						testutils.HistogramBucketValue(t, promRegistry, "eventstore_batch_size_events",
							nil, 10))
				})
			})

			t.Run("MultipleEvents/Pageviews/Custom/OutboundLinkClick", func(t *testing.T) {
				forEachEventDb(t, cfg, func(t *testing.T, store Service, db eventdb.Service, promRegistry *prometheus.Registry) {
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
						err := store.StorePageView(context.Background(), &event.PageView{
							Timestamp: eventTime,
							PageUri:   testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
							Session:   session,
						})
						require.NoError(t, err)

						// Custom event associated to the same session.
						eventTime = time.Now().UTC().Round(time.Second)
						err = store.StoreCustom(context.Background(), &event.Custom{
							Timestamp: eventTime,
							PageUri:   testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
							Session:   session,
							Name:      "foo",
							Keys:      []string{},
							Values:    []string{},
						})
						require.NoError(t, err)

						eventTime = time.Now().UTC().Round(time.Second)
						err = store.StoreOutboundLinkClick(context.Background(), &event.OutboundLinkClick{
							Timestamp: eventTime,
							PageUri:   testutils.Must(uri.Parse)("http://mywebsite.localhost/"),
							Session:   session,
							Link:      testutils.Must(uri.Parse)("http://example.com"),
						})
						require.NoError(t, err)

						eventTime = time.Now().UTC().Round(time.Second)
						err = store.StoreFileDownload(context.Background(), &event.FileDownload{
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
						row, err := db.Query(
							context.Background(),
							"SELECT COUNT(*) FROM prisme.pageviews WHERE timestamp >= ? AND session_uuid IN (SELECT session_uuid FROM prisme.sessions WHERE visitor_id = 'multipleEventsTestCase')",
							testStartTime.Unix(),
						)
						require.NoError(t, err)
						require.True(t, row.Next())

						var pageviewsCount uint64
						err = row.Scan(&pageviewsCount)
						require.NoError(t, err)
						require.Equal(t, uint64(sessionsCount), pageviewsCount)
					}

					// Ensure custom events are stored.
					{
						row, err := db.Query(
							context.Background(),
							"SELECT COUNT(*) FROM prisme.events_custom WHERE timestamp >= ? AND session_uuid IN (SELECT session_uuid FROM prisme.sessions WHERE visitor_id = 'multipleEventsTestCase')",
							testStartTime.Unix(),
						)
						require.NoError(t, err)
						require.True(t, row.Next())

						var customEventsCount uint64
						err = row.Scan(&customEventsCount)
						require.NoError(t, err)
						require.Equal(t, uint64(sessionsCount), customEventsCount)
					}

					// Check metrics.
					{
						require.Equal(t, float64(0),
							testutils.CounterValue(t, promRegistry, "eventstore_batch_dropped_total",
								nil))
						require.Equal(t, float64(0),
							testutils.CounterValue(t, promRegistry, "eventstore_batch_retry_total",
								nil))
						require.Equal(t, float64(sessionsCount*int(maxEventKind)),
							testutils.CounterValue(t, promRegistry, "eventstore_events_total",
								nil))
						require.Equal(t, float64(0),
							testutils.CounterValue(t, promRegistry, "eventstore_ring_buffers_dropped_events_total",
								nil))
						require.Greater(t,
							testutils.HistogramSumValue(t, promRegistry, "eventstore_send_batch_duration_seconds", nil),
							float64(0))
						require.Equal(t, float64(sessionsCount*int(maxEventKind)),
							testutils.HistogramSumValue(t, promRegistry, "eventstore_batch_size_events",
								nil))
					}
				})
			})

			t.Run("RingBufferDroppedEvents", func(t *testing.T) {
				cfg := Config{
					MaxBatchSize:      1_000,
					MaxBatchTimeout:   10 * time.Millisecond,
					RingBuffersFactor: 1,
				}

				forEachEventDb(t, cfg, func(t *testing.T, store Service, db eventdb.Service, promRegistry *prometheus.Registry) {
					// Send hundreds of event without pause.
					for i := 0; i < 10_000; i++ {
						eventTime := time.Now().UTC().Round(time.Second)
						err := store.StorePageView(context.Background(), &event.PageView{
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
						require.Greater(t,
							testutils.CounterValue(t, promRegistry, "eventstore_ring_buffers_dropped_events_total", nil),
							float64(0))
					}
				})
			})
		})
	}
}
