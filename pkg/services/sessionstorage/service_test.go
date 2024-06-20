package sessionstorage

import (
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func randomDeviceId() string {
	return fmt.Sprintf("%X", rand.Int63())
}

func TestService(t *testing.T) {
	logger := log.NewLogger("sessionstorage_test", io.Discard, true)
	cfg := Config{
		gcInterval:         10 * time.Second,
		sessionInactiveTtl: 24 * time.Hour,
	}

	t.Run("GetSession", func(t *testing.T) {
		t.Run("NonExistent", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			storedSession, ok := service.GetSession("...")
			require.False(t, ok)
			require.Equal(t, storedSession, event.Session{})

			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
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
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
		})

		t.Run("Expired", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, Config{
				gcInterval:         30 * time.Millisecond,
				sessionInactiveTtl: 30 * time.Millisecond,
			}, promRegistry)

			deviceId := randomDeviceId()
			session := event.Session{
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			service.InsertSession(deviceId, session)

			time.Sleep(35 * time.Millisecond)

			storedSession, ok := service.GetSession(deviceId)
			require.False(t, ok)
			require.Equal(t, storedSession, event.Session{})

			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "overwritten"}))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "expired"}))
			require.Equal(t, float64(1),
				testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
		})
	})

	t.Run("InsertSession", func(t *testing.T) {
		t.Run("NonExistent", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := randomDeviceId()
			session := event.Session{
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			service.InsertSession(deviceId, session)

			storedSession, ok := service.GetSession(deviceId)
			require.True(t, ok)
			require.Equal(t, storedSession, session)

			require.Equal(t, float64(1),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
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
			service := ProvideService(logger, cfg, promRegistry)

			deviceId := randomDeviceId()
			sessionA := event.Session{
				VisitorId:     "prisme_XXX",
				PageviewCount: 1,
			}

			service.InsertSession(deviceId, sessionA)

			sessionB := sessionA
			sessionB.VisitorId = "prisme_YYY"

			// Overwrite session A.
			service.InsertSession(deviceId, sessionB)

			storedSession, ok := service.GetSession(deviceId)
			require.True(t, ok)
			require.Equal(t, storedSession, sessionB)

			require.Equal(t, float64(1),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
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

	t.Run("IncSessionPageviewCount", func(t *testing.T) {
		promRegistry := prometheus.NewRegistry()
		service := ProvideService(logger, cfg, promRegistry)

		deviceId := randomDeviceId()
		sessionV1 := event.Session{
			VisitorId:     "prisme_XXX",
			PageviewCount: 1,
		}

		service.InsertSession(deviceId, sessionV1)

		sessionV2, ok := service.IncSessionPageviewCount(deviceId)
		require.True(t, ok)
		require.Equal(t, sessionV1.VisitorId, sessionV2.VisitorId)

		require.Equal(t, sessionV1.PageviewCount+1, sessionV2.PageviewCount)

		require.Equal(t, float64(1),
			testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
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
}
