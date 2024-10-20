package sessionstorage

import (
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	logger := log.NewLogger("sessionstorage_test", io.Discard, true)
	cfg := Config{
		gcInterval:            10 * time.Second,
		sessionInactiveTtl:    24 * time.Hour,
		maxSessionsPerVisitor: 64,
	}

	mustParseUri := testutils.Must(uri.Parse)

	t.Run("InsertSession", func(t *testing.T) {
		t.Run("NonExistent", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry).(*service)

			deviceId := rand.Uint64()
			pageUri := mustParseUri("https://example.com")
			session := event.Session{
				PageUri:       pageUri,
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			ok := service.InsertSession(deviceId, session)
			require.True(t, ok)

			entry := service.getValidSessionEntry(deviceId, pageUri.Path())
			require.NotNil(t, entry)
			require.Equal(t, entry.Session, session)

			require.Equal(t, float64(1),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "expired"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
		})

		t.Run("Existent", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry).(*service)

			deviceId := rand.Uint64()
			pageUri := mustParseUri("https://example.com")
			sessionA := event.Session{
				PageUri:       pageUri,
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			ok := service.InsertSession(deviceId, sessionA)
			require.True(t, ok)

			sessionB := sessionA
			sessionB.VisitorId = "prisme_YYY"

			// Add another session on same path.
			ok = service.InsertSession(deviceId, sessionB)
			require.True(t, ok)

			// get session returns first matching session, here it is session A.
			entry := service.getValidSessionEntry(deviceId, pageUri.Path())
			require.NotNil(t, entry)
			require.Equal(t, entry.Session, sessionA)

			require.Equal(t, float64(2),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))
			require.Equal(t, float64(2),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "expired"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
		})

		t.Run("TooManySession", func(t *testing.T) {
			testCfg := cfg
			testCfg.maxSessionsPerVisitor = 1

			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, testCfg, promRegistry).(*service)

			deviceId := rand.Uint64()
			pageUri := mustParseUri("https://example.com")
			sessionA := event.Session{
				PageUri:       pageUri,
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			ok := service.InsertSession(deviceId, sessionA)
			require.True(t, ok)

			sessionB := sessionA
			sessionB.VisitorId = "prisme_YYY"

			ok = service.InsertSession(deviceId, sessionB)
			require.False(t, ok)
		})
	})

	t.Run("AddPageview", func(t *testing.T) {
		t.Run("WrongPath", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := rand.Uint64()
			pageUri := mustParseUri("https://example.com")
			sessionV1 := event.Session{
				PageUri:       pageUri,
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			ok := service.InsertSession(deviceId, sessionV1)
			require.True(t, ok)

			// Referrer doesn't match page uri of created session.
			referrer := event.ReferrerUri{Uri: mustParseUri("https://example.com/bar")}
			pageUri = mustParseUri("https://example.com/foo")

			sessionV2, ok := service.AddPageview(deviceId, referrer, pageUri)
			require.False(t, ok)
			require.NotEqual(t, sessionV1.VisitorId, sessionV2.VisitorId)

			require.Equal(t, float64(1),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "overwritten"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "expired"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
		})

		t.Run("RightPath", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := rand.Uint64()
			pageUri := mustParseUri("https://example.com")
			sessionV1 := event.Session{
				PageUri:       pageUri,
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			ok := service.InsertSession(deviceId, sessionV1)
			require.True(t, ok)

			referrer := event.ReferrerUri{Uri: pageUri}
			pageUri = mustParseUri("https://example.com/foo")

			sessionV2, ok := service.AddPageview(deviceId, referrer, pageUri)
			require.True(t, ok)
			require.Equal(t, sessionV1.VisitorId, sessionV2.VisitorId)

			require.Equal(t, sessionV1.PageviewCount+1, sessionV2.PageviewCount)

			require.Equal(t, float64(1),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "overwritten"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "expired"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
		})
	})

	t.Run("IdentifySession", func(t *testing.T) {
		t.Run("RightPath", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := rand.Uint64()
			pageUri := mustParseUri("https://example.com")
			session := event.Session{
				PageUri:       pageUri,
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			ok := service.InsertSession(deviceId, session)
			require.True(t, ok)

			identifiedSession, ok := service.IdentifySession(deviceId, session.PageUri, "prisme_YYY")
			require.True(t, ok)
			require.Equal(t, identifiedSession.VisitorId, "prisme_YYY")
			require.Equal(t, identifiedSession.PageUri, session.PageUri)
		})

		t.Run("WrongPath", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := rand.Uint64()
			pageUri := mustParseUri("https://example.com")
			session := event.Session{
				PageUri:       pageUri,
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			ok := service.InsertSession(deviceId, session)
			require.True(t, ok)

			identifiedSession, ok := service.IdentifySession(deviceId, mustParseUri("https://example.com/foo"), "prisme_YYY")
			require.False(t, ok)
			require.NotEqual(t, identifiedSession.VisitorId, "prisme_YYY")
			require.NotEqual(t, identifiedSession.PageUri, session.PageUri)
		})
	})

	t.Run("WaitSession", func(t *testing.T) {
		t.Run("Timeout", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := rand.Uint64()
			pageUri := mustParseUri("https://example.com")

			// No session wait.
			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))

			// Check session wait metric in parallel.
			go func() {
				time.Sleep(5 * time.Millisecond)

				// A single session wait.
				require.Equal(t, float64(1),
					testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))
			}()

			// Wait for session.
			now := time.Now()
			session, found := service.WaitSession(deviceId, pageUri, 10*time.Millisecond)
			require.False(t, found)
			require.Equal(t, event.Session{}, session)
			require.WithinDuration(t, now.Add(10*time.Millisecond), time.Now(), 3*time.Millisecond)

			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "overwritten"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "expired"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
		})

		t.Run("Created", func(t *testing.T) {
			t.Run("RightPath", func(t *testing.T) {
				promRegistry := prometheus.NewRegistry()
				service := ProvideService(logger, cfg, promRegistry)

				deviceId := rand.Uint64()
				pageUri := mustParseUri("https://example.com")
				session := event.Session{
					PageUri:       pageUri,
					VisitorId:     "prisme_XXX",
					PageviewCount: 1,
				}

				// No session wait.
				require.Equal(t, float64(0),
					testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))

				// Insert session in parallel.
				go func() {
					require.Equal(t, float64(0),
						testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
					require.Equal(t, float64(1),
						testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))
					require.Equal(t, float64(0),
						testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
							prometheus.Labels{"type": "inserted"}))
					require.Equal(t, float64(0),
						testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
							prometheus.Labels{"type": "overwritten"}))
					require.Equal(t, float64(0),
						testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
							prometheus.Labels{"type": "expired"}))
					require.Equal(t, float64(0),
						testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))

					time.Sleep(5 * time.Millisecond)

					// A single session wait.
					require.Equal(t, float64(1),
						testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))

					ok := service.InsertSession(deviceId, session)
					require.True(t, ok)
				}()

				// Wait for session.
				now := time.Now()
				actualSession, found := service.WaitSession(deviceId, pageUri, 10*time.Millisecond)
				require.True(t, found)
				require.Equal(t, session, actualSession)
				require.WithinDuration(t, now.Add(5*time.Millisecond), time.Now(), 3*time.Millisecond)

				require.Equal(t, float64(1),
					testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
				require.Equal(t, float64(0),
					testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))
				require.Equal(t, float64(1),
					testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
						prometheus.Labels{"type": "inserted"}))
				require.Equal(t, float64(0),
					testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
						prometheus.Labels{"type": "overwritten"}))
				require.Equal(t, float64(0),
					testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
						prometheus.Labels{"type": "expired"}))
				require.Equal(t, float64(0),
					testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
			})

			t.Run("WrongPath", func(t *testing.T) {
				promRegistry := prometheus.NewRegistry()
				service := ProvideService(logger, cfg, promRegistry)

				deviceId := rand.Uint64()
				session := event.Session{
					PageUri:       mustParseUri("https://example.com"),
					VisitorId:     "prisme_XXX",
					PageviewCount: 1,
				}

				// No session wait.
				require.Equal(t, float64(0),
					testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))

				// Insert session in parallel.
				go func() {
					require.Equal(t, float64(0),
						testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
					require.Equal(t, float64(1),
						testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))
					require.Equal(t, float64(0),
						testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
							prometheus.Labels{"type": "inserted"}))
					require.Equal(t, float64(0),
						testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
							prometheus.Labels{"type": "overwritten"}))
					require.Equal(t, float64(0),
						testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
							prometheus.Labels{"type": "expired"}))
					require.Equal(t, float64(0),
						testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))

					time.Sleep(5 * time.Millisecond)

					// A single session wait.
					require.Equal(t, float64(1),
						testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))

					ok := service.InsertSession(deviceId, session)
					require.True(t, ok)
				}()

				// Wait for session with /foo path.
				now := time.Now()
				actualSession, found := service.WaitSession(deviceId, mustParseUri("https://example.com/foo"), 10*time.Millisecond)
				require.False(t, found)
				require.NotEqual(t, session, actualSession)
				require.WithinDuration(t, now.Add(10*time.Millisecond), time.Now(), 3*time.Millisecond)

				require.Equal(t, float64(1),
					testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
				require.Equal(t, float64(0),
					testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))
				require.Equal(t, float64(1),
					testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
						prometheus.Labels{"type": "inserted"}))
				require.Equal(t, float64(0),
					testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
						prometheus.Labels{"type": "overwritten"}))
				require.Equal(t, float64(0),
					testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
						prometheus.Labels{"type": "expired"}))
				require.Equal(t, float64(0),
					testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))

			})
		})

		t.Run("AlreadyExists", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := rand.Uint64()
			pageUri := mustParseUri("https://example.com")
			session := event.Session{
				PageUri:       pageUri,
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			// Insert session.
			ok := service.InsertSession(deviceId, session)
			require.True(t, ok)

			now := time.Now()
			actualSession, found := service.WaitSession(deviceId, pageUri, 10*time.Millisecond)
			require.True(t, found)
			require.Equal(t, session, actualSession)
			require.WithinDuration(t, now, time.Now(), 3*time.Millisecond)

			require.Equal(t, float64(1),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "overwritten"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "expired"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
		})
	})
}
