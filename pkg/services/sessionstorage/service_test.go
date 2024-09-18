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
		gcInterval:         10 * time.Second,
		sessionInactiveTtl: 24 * time.Hour,
	}

	mustParseUri := testutils.Must(uri.Parse)

	t.Run("InsertSession", func(t *testing.T) {
		t.Run("NonExistent", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry).(*service)

			deviceId := rand.Uint64()
			session := event.Session{
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			service.InsertSession(deviceId, session)

			sessionEntry, ok := service.getSessionEntry(deviceId)
			require.True(t, ok)
			require.Equal(t, sessionEntry.session, session)

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
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
		})

		t.Run("Existent", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry).(*service)

			deviceId := rand.Uint64()
			sessionA := event.Session{
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			service.InsertSession(deviceId, sessionA)

			sessionB := sessionA
			sessionB.VisitorId = "prisme_YYY"

			// Overwrite session A.
			service.InsertSession(deviceId, sessionB)

			sessionEntry, ok := service.getSessionEntry(deviceId)
			require.True(t, ok)
			require.Equal(t, sessionEntry.session, sessionB)

			require.Equal(t, float64(1),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))
			require.Equal(t, float64(2),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "overwritten"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "expired"}))
			require.Equal(t, float64(1),
				testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
		})
	})

	t.Run("AddPageview", func(t *testing.T) {
		t.Run("Duplicate", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := rand.Uint64()
			session := event.Session{
				PageUri:       mustParseUri("https://example.com"),
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			service.InsertSession(deviceId, session)

			result, ok := service.AddPageview(deviceId, session.PageUri)
			require.True(t, ok)
			require.True(t, result.DuplicatePageview)
			require.Equal(t, session.VisitorId, result.Session.VisitorId)
			require.Equal(t, session.PageviewCount, result.Session.PageviewCount)

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
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
		})
		t.Run("Unique", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := rand.Uint64()
			session := event.Session{
				PageUri:       mustParseUri("https://example.com"),
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			service.InsertSession(deviceId, session)

			result, ok := service.AddPageview(deviceId, mustParseUri("https://example.com/foo"))
			require.True(t, ok)
			require.False(t, result.DuplicatePageview)
			require.Equal(t, session.VisitorId, result.Session.VisitorId)
			require.Equal(t, session.PageviewCount+1, result.Session.PageviewCount)

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
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
		})
	})

	t.Run("WaitSession", func(t *testing.T) {
		t.Run("Timeout", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := rand.Uint64()

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
			session, found := service.WaitSession(deviceId, 10*time.Millisecond)
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
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
		})

		t.Run("Created", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := rand.Uint64()
			session := event.Session{
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
				require.Equal(t, float64(0),
					testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))

				time.Sleep(5 * time.Millisecond)

				// A single session wait.
				require.Equal(t, float64(1),
					testutils.GaugeValue(t, promRegistry, "sessionstorage_sessions_wait", nil))

				service.InsertSession(deviceId, session)
			}()

			// Wait for session.
			now := time.Now()
			actualSession, found := service.WaitSession(deviceId, 10*time.Millisecond)
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
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
		})

		t.Run("AlreadyExists", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := rand.Uint64()
			session := event.Session{
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			// Insert session.
			service.InsertSession(deviceId, session)

			now := time.Now()
			actualSession, found := service.WaitSession(deviceId, 10*time.Millisecond)
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
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
		})
	})
}
