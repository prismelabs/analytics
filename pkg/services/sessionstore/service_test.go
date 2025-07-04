package sessionstore

import (
	"fmt"
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	logger := log.NewLogger("sessionstore_test", io.Discard, true)
	cfg := Config{
		gcInterval:             10 * time.Second,
		sessionInactiveTtl:     24 * time.Hour,
		deviceExpiryPercentile: 0, // Collect session as soon it expire.
		maxSessionsPerVisitor:  64,
	}

	getValidSessionEntry := func(s *service, deviceId uint64, latestPath string) sessionEntry {
		s.mu.Lock()
		defer s.mu.Unlock()
		entry := s.getValidSessionEntry(deviceId, latestPath)
		return *entry
	}

	mustParseUri := testutils.Must(uri.Parse)
	mustUuidV7 := testutils.MustNoArg(uuid.NewV7)
	mustParseReferrerUri := testutils.Must(event.ParseReferrerUri)

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

			entry := getValidSessionEntry(service, deviceId, pageUri.Path())
			require.NotNil(t, entry)
			require.Equal(t, entry.Session, session)

			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))
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
			entry := getValidSessionEntry(service, deviceId, pageUri.Path())
			require.NotNil(t, entry)
			require.Equal(t, entry.Session, sessionA)

			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(2),
				testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))
		})

		t.Run("TooManySession", func(t *testing.T) {
			cfg := cfg
			cfg.maxSessionsPerVisitor = 1

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
			referrer := mustParseReferrerUri([]byte(mustParseUri("https://example.com/bar").String()))
			pageUri = mustParseUri("https://example.com/foo")

			sessionV2, ok := service.AddPageview(deviceId, referrer, pageUri)
			require.False(t, ok)
			require.NotEqual(t, sessionV1.VisitorId, sessionV2.VisitorId)

			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))
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

			referrer := mustParseReferrerUri([]byte(mustParseUri("https://example.com/").String()))
			pageUri = mustParseUri("https://example.com/foo")

			sessionV2, ok := service.AddPageview(deviceId, referrer, pageUri)
			require.True(t, ok)
			require.Equal(t, sessionV1.VisitorId, sessionV2.VisitorId)
			require.Equal(t, sessionV1.PageviewCount+1, sessionV2.PageviewCount)

			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))
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
				testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))

			// Check session wait metric in parallel.
			go func() {
				time.Sleep(5 * time.Millisecond)

				// A single session wait.
				require.Equal(t, float64(1),
					testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))
			}()

			// Wait for session.
			now := time.Now()
			session, found := service.WaitSession(deviceId, pageUri, 10*time.Millisecond)
			require.False(t, found)
			require.Equal(t, event.Session{}, session)
			require.WithinDuration(t, now.Add(10*time.Millisecond), time.Now(), 3*time.Millisecond)

			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))
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
					testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))

				// Insert session in parallel.
				go func() {
					require.Equal(t, float64(1),
						testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))
					require.Equal(t, float64(1),
						testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
							prometheus.Labels{"type": "inserted"}))
					require.Equal(t, float64(1),
						testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
							prometheus.Labels{"type": "inserted"}))
					require.Equal(t, float64(0),
						testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))

					time.Sleep(5 * time.Millisecond)

					// A single session wait.
					require.Equal(t, float64(1),
						testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))

					ok := service.InsertSession(deviceId, session)
					require.True(t, ok)
				}()

				// Wait for session.
				now := time.Now()
				actualSession, found := service.WaitSession(deviceId, pageUri, 10*time.Millisecond)
				require.True(t, found)
				require.Equal(t, session, actualSession)
				require.WithinDuration(t, now.Add(5*time.Millisecond), time.Now(), 3*time.Millisecond)

				require.Equal(t, float64(0),
					testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))
				require.Equal(t, float64(1),
					testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
						prometheus.Labels{"type": "inserted"}))
				require.Equal(t, float64(1),
					testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
						prometheus.Labels{"type": "inserted"}))
				require.Equal(t, float64(0),
					testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))
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
					testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))

				// Insert session in parallel.
				go func() {
					require.Equal(t, float64(1),
						testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))
					require.Equal(t, float64(1),
						testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
							prometheus.Labels{"type": "inserted"}))
					require.Equal(t, float64(1),
						testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
							prometheus.Labels{"type": "inserted"}))
					require.Equal(t, float64(0),
						testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))

					time.Sleep(5 * time.Millisecond)

					ok := service.InsertSession(deviceId, session)
					require.True(t, ok)
				}()

				// Wait for session with /foo path.
				now := time.Now()
				actualSession, found := service.WaitSession(deviceId, mustParseUri("https://example.com/foo"), 10*time.Millisecond)
				require.False(t, found)
				require.NotEqual(t, session, actualSession)
				require.WithinDuration(t, now.Add(10*time.Millisecond), time.Now(), 3*time.Millisecond)

				require.Equal(t, float64(0),
					testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))
				require.Equal(t, float64(1),
					testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
						prometheus.Labels{"type": "inserted"}))
				require.Equal(t, float64(2),
					testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
						prometheus.Labels{"type": "inserted"}))
				require.Equal(t, float64(0),
					testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))
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

			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstore_sessions_wait", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))
		})
	})

	t.Run("GC", func(t *testing.T) {
		cfg := Config{
			gcInterval:             10 * time.Millisecond,
			sessionInactiveTtl:     2 * time.Second,
			deviceExpiryPercentile: 0, // Collect session as soon it expire.
			maxSessionsPerVisitor:  64,
		}

		t.Run("SingleDevice", func(t *testing.T) {
			t.Run("SingleSession", func(t *testing.T) {
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

				entry := getValidSessionEntry(service, deviceId, pageUri.Path())
				require.NotNil(t, entry)
				require.Equal(t, entry.Session, session)

				require.Equal(t, float64(0),
					testutils.GaugeValue(t, promRegistry, "sessionstore_gc_cycles_total", nil))

				// Wait for GC.
				time.Sleep(cfg.sessionInactiveTtl + cfg.gcInterval)

				require.GreaterOrEqual(t,
					testutils.CounterValue(t, promRegistry, "sessionstore_gc_cycles_total", nil),
					float64(10))
				require.GreaterOrEqual(t,
					testutils.HistogramSumValue(t, promRegistry, "sessionstore_gc_cycles_duration_ms", nil),
					float64(0))
				require.Equal(t, float64(1),
					testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
						prometheus.Labels{"type": "inserted"}))
				require.Equal(t, float64(1),
					testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
						prometheus.Labels{"type": "deleted"}))
				require.Equal(t, float64(1),
					testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
						prometheus.Labels{"type": "inserted"}))
				require.Equal(t, float64(1),
					testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
						prometheus.Labels{"type": "expired"}))
				require.Equal(t, float64(1),
					testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))

				service.mu.Lock()
				_, ok = service.devices[deviceId]
				require.False(t, ok)
				service.mu.Unlock()
			})

			t.Run("MultipleSessions", func(t *testing.T) {
				t.Run("SingleExpired", func(t *testing.T) {
					t.Run("p(0)", func(t *testing.T) {
						promRegistry := prometheus.NewRegistry()
						service := ProvideService(logger, cfg, promRegistry).(*service)

						deviceId := rand.Uint64()
						activeSessions := sync.Map{}

						// Insert 10 sessions.
						for i := 1; i <= 10; i++ {
							pageUri := mustParseUri(fmt.Sprintf("https://example.com/%v", i))
							session := event.Session{
								SessionUuid:   mustUuidV7(),
								PageUri:       pageUri,
								VisitorId:     "prisme_XXX",
								PageviewCount: uint16(i),
							}

							ok := service.InsertSession(deviceId, session)
							require.True(t, ok)

							entry := getValidSessionEntry(service, deviceId, pageUri.Path())
							require.NotNil(t, entry)
							require.Equal(t, entry.Session, session)

							// Add pageview for all except 5th session.
							if i != 5 {
								newPageUri := mustParseUri(pageUri.String() + "/foo")
								referrerUri := mustParseReferrerUri([]byte(session.PageUri.String()))
								go func() {
									time.Sleep(cfg.sessionInactiveTtl / 2)
									session, ok := service.AddPageview(deviceId, referrerUri, newPageUri)
									require.True(t, ok)
									activeSessions.Store(session.SessionUuid, session)
								}()
							}
						}

						// Wait for GC.
						time.Sleep(cfg.sessionInactiveTtl + cfg.gcInterval)

						// Metrics.
						require.GreaterOrEqual(t,
							testutils.CounterValue(t, promRegistry, "sessionstore_gc_cycles_total", nil),
							float64(10))
						require.GreaterOrEqual(t,
							testutils.HistogramSumValue(t, promRegistry, "sessionstore_gc_cycles_duration_ms", nil),
							float64(0))
						require.Equal(t, float64(1),
							testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
								prometheus.Labels{"type": "inserted"}))
						require.Equal(t, float64(0),
							testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
								prometheus.Labels{"type": "deleted"}))
						require.Equal(t, float64(10),
							testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
								prometheus.Labels{"type": "inserted"}))
						require.Equal(t, float64(1),
							testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
								prometheus.Labels{"type": "expired"}))
						require.Equal(t, float64(5),
							testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))

						service.mu.Lock()
						device := service.devices[deviceId]
						for _, entry := range device.entries {
							expected, ok := activeSessions.Load(entry.Session.SessionUuid)
							require.True(t, ok)
							require.Equal(t, expected, entry.Session)
						}
						service.mu.Unlock()
					})

					t.Run("p(100)", func(t *testing.T) {
						cfg := cfg
						cfg.deviceExpiryPercentile = 100

						promRegistry := prometheus.NewRegistry()
						service := ProvideService(logger, cfg, promRegistry).(*service)

						deviceId := rand.Uint64()

						// Insert 10 sessions.
						for i := 1; i <= 10; i++ {
							pageUri := mustParseUri(fmt.Sprintf("https://example.com/%v", i))
							session := event.Session{
								SessionUuid:   mustUuidV7(),
								PageUri:       pageUri,
								VisitorId:     "prisme_XXX",
								PageviewCount: 1,
							}

							ok := service.InsertSession(deviceId, session)
							require.True(t, ok)

							entry := getValidSessionEntry(service, deviceId, pageUri.Path())
							require.NotNil(t, entry)
							require.Equal(t, entry.Session, session)

							// Last session will expire later.
							if i == 10 {
								newPageUri := mustParseUri(pageUri.String() + "/foo")
								referrerUri := mustParseReferrerUri([]byte(session.PageUri.String()))
								go func() {
									time.Sleep(cfg.sessionInactiveTtl / 2)
									_, ok := service.AddPageview(deviceId, referrerUri, newPageUri)
									require.True(t, ok)
								}()
							}
						}

						// Wait for GC.
						time.Sleep(cfg.sessionInactiveTtl + cfg.gcInterval)

						// Metrics.
						// No session has been collected has device sessions are collected all at once.
						require.GreaterOrEqual(t,
							testutils.CounterValue(t, promRegistry, "sessionstore_gc_cycles_total", nil),
							float64(10))
						require.GreaterOrEqual(t,
							testutils.HistogramSumValue(t, promRegistry, "sessionstore_gc_cycles_duration_ms", nil),
							float64(0))
						require.Equal(t, float64(1),
							testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
								prometheus.Labels{"type": "inserted"}))
						require.Equal(t, float64(0),
							testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
								prometheus.Labels{"type": "deleted"}))
						require.Equal(t, float64(10),
							testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
								prometheus.Labels{"type": "inserted"}))
						require.Equal(t, float64(0),
							testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
								prometheus.Labels{"type": "expired"}))
						require.Equal(t, float64(0),
							testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))

						// Wait for GC.
						time.Sleep(cfg.sessionInactiveTtl + cfg.gcInterval)

						// All device's sessions has been collected.
						require.GreaterOrEqual(t,
							testutils.CounterValue(t, promRegistry, "sessionstore_gc_cycles_total", nil),
							float64(10))
						require.GreaterOrEqual(t,
							testutils.HistogramSumValue(t, promRegistry, "sessionstore_gc_cycles_duration_ms", nil),
							float64(0))
						require.Equal(t, float64(1),
							testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
								prometheus.Labels{"type": "inserted"}))
						require.Equal(t, float64(1),
							testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
								prometheus.Labels{"type": "deleted"}))
						require.Equal(t, float64(10),
							testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
								prometheus.Labels{"type": "inserted"}))
						require.Equal(t, float64(10),
							testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
								prometheus.Labels{"type": "expired"}))
						require.Equal(t, float64(10+1),
							testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))
					})
				})

				t.Run("MultipleExpired", func(t *testing.T) {
					promRegistry := prometheus.NewRegistry()
					service := ProvideService(logger, cfg, promRegistry).(*service)

					deviceId := rand.Uint64()

					// Insert 10 sessions.
					for i := 1; i <= 10; i++ {
						pageUri := mustParseUri(fmt.Sprintf("https://example.com/%v", i))
						session := event.Session{
							SessionUuid:   mustUuidV7(),
							PageUri:       pageUri,
							VisitorId:     "prisme_XXX",
							PageviewCount: uint16(i),
						}

						ok := service.InsertSession(deviceId, session)
						require.True(t, ok)

						entry := getValidSessionEntry(service, deviceId, pageUri.Path())
						require.NotNil(t, entry)
						require.Equal(t, entry.Session, session)

						// Add pageview for all except 5th and 6th sessions.
						if i != 5 && i != 6 {
							newPageUri := mustParseUri(pageUri.String() + "/foo")
							referrerUri := mustParseReferrerUri([]byte(session.PageUri.String()))
							go func() {
								time.Sleep(cfg.sessionInactiveTtl / 2)
								_, ok := service.AddPageview(deviceId, referrerUri, newPageUri)
								require.True(t, ok)
							}()
						}
					}

					// Wait for GC.
					time.Sleep(cfg.sessionInactiveTtl + cfg.gcInterval)

					require.GreaterOrEqual(t,
						testutils.CounterValue(t, promRegistry, "sessionstore_gc_cycles_total", nil),
						float64(10))
					require.GreaterOrEqual(t,
						testutils.HistogramSumValue(t, promRegistry, "sessionstore_gc_cycles_duration_ms", nil),
						float64(0))
					require.Equal(t, float64(1),
						testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
							prometheus.Labels{"type": "inserted"}))
					require.Equal(t, float64(0),
						testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
							prometheus.Labels{"type": "deleted"}))
					require.Equal(t, float64(10),
						testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
							prometheus.Labels{"type": "inserted"}))
					require.Equal(t, float64(2),
						testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
							prometheus.Labels{"type": "expired"}))
					require.Equal(t, float64(5+6),
						testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))
				})

				t.Run("AllExpired", func(t *testing.T) {
					promRegistry := prometheus.NewRegistry()
					service := ProvideService(logger, cfg, promRegistry).(*service)

					deviceId := rand.Uint64()

					// Insert 10 sessions.
					for i := 1; i <= 10; i++ {
						pageUri := mustParseUri(fmt.Sprintf("https://example.com/%v", i))
						session := event.Session{
							SessionUuid:   mustUuidV7(),
							PageUri:       pageUri,
							VisitorId:     "prisme_XXX",
							PageviewCount: uint16(i),
						}

						ok := service.InsertSession(deviceId, session)
						require.True(t, ok)

						entry := getValidSessionEntry(service, deviceId, pageUri.Path())
						require.NotNil(t, entry)
						require.Equal(t, entry.Session, session)
					}

					// Wait for GC.
					time.Sleep(cfg.sessionInactiveTtl + cfg.gcInterval)

					// Metrics.
					require.GreaterOrEqual(t,
						testutils.CounterValue(t, promRegistry, "sessionstore_gc_cycles_total", nil),
						float64(10))
					require.GreaterOrEqual(t,
						testutils.HistogramSumValue(t, promRegistry, "sessionstore_gc_cycles_duration_ms", nil),
						float64(0))
					require.Equal(t, float64(1),
						testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
							prometheus.Labels{"type": "inserted"}))
					require.Equal(t, float64(1),
						testutils.CounterValue(t, promRegistry, "sessionstore_devices_total",
							prometheus.Labels{"type": "deleted"}))
					require.Equal(t, float64(10),
						testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
							prometheus.Labels{"type": "inserted"}))
					require.Equal(t, float64(10),
						testutils.CounterValue(t, promRegistry, "sessionstore_sessions_total",
							prometheus.Labels{"type": "expired"}))
					require.Equal(t, float64(10+9+8+7+6+5+4+3+2+1),
						testutils.HistogramSumValue(t, promRegistry, "sessionstore_sessions_pageviews", nil))

					service.mu.Lock()
					_, ok := service.devices[deviceId]
					require.False(t, ok)
					service.mu.Unlock()
				})
			})
		})
	})
}
