package sessionstorage

import (
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

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
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total", nil))
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

			session := event.Session{
				VisitorId: "prisme_XXX",
				Pageviews: 1,
			}

			upserted := service.UpsertSession(session)
			require.True(t, upserted)

			time.Sleep(35 * time.Millisecond)

			storedSession, ok := service.GetSession(session.VisitorId)
			require.False(t, ok)
			require.Equal(t, storedSession, event.Session{})

			require.Equal(t, float64(0),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "expired"}))
			require.Equal(t, float64(1),
				testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
		})
	})

	t.Run("UpsertSession", func(t *testing.T) {
		t.Run("NonExistent", func(t *testing.T) {
			promRegistry := prometheus.NewRegistry()
			service := ProvideService(logger, cfg, promRegistry)

			session := event.Session{
				VisitorId: "prisme_XXX",
				Pageviews: 1,
			}

			upserted := service.UpsertSession(session)
			require.True(t, upserted)

			storedSession, ok := service.GetSession(session.VisitorId)
			require.True(t, ok)
			require.Equal(t, storedSession, session)

			require.Equal(t, float64(1),
				testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
			require.Equal(t, float64(1),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "inserted"}))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
					prometheus.Labels{"type": "expired"}))
			require.Equal(t, float64(0),
				testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
			require.Equal(t, float64(0),
				testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
		})

		t.Run("Existent", func(t *testing.T) {
			t.Run("NewerVersion", func(t *testing.T) {
				promRegistry := prometheus.NewRegistry()
				service := ProvideService(logger, cfg, promRegistry)

				sessionV1 := event.Session{
					VisitorId: "prisme_XXX",
					Pageviews: 1,
				}

				upserted := service.UpsertSession(sessionV1)
				require.True(t, upserted)

				sessionV2 := sessionV1
				sessionV2.Pageviews++

				upserted = service.UpsertSession(sessionV2)
				require.True(t, upserted)

				// Upserting v1 should fail.
				upserted = service.UpsertSession(sessionV1)
				require.False(t, upserted)

				storedSession, ok := service.GetSession(sessionV1.VisitorId)
				require.True(t, ok)
				require.Equal(t, storedSession, sessionV2) // session v2 overwrite session v1

				require.Equal(t, float64(1),
					testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
				require.Equal(t, float64(1),
					testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
						prometheus.Labels{"type": "inserted"}))
				require.Equal(t, float64(0),
					testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
						prometheus.Labels{"type": "expired"}))
				require.Equal(t, float64(0),
					testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
				require.Equal(t, float64(0),
					testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
			})

			t.Run("OlderVersion", func(t *testing.T) {
				promRegistry := prometheus.NewRegistry()
				service := ProvideService(logger, cfg, promRegistry)

				sessionV1 := event.Session{
					VisitorId: "prisme_XXX",
					Pageviews: 1,
				}

				upserted := service.UpsertSession(sessionV1)
				require.True(t, upserted)

				sessionV2 := sessionV1
				sessionV2.Pageviews++

				upserted = service.UpsertSession(sessionV2)
				require.True(t, upserted)

				// Upserting v1 should fail.
				upserted = service.UpsertSession(sessionV1)
				require.False(t, upserted)

				storedSession, ok := service.GetSession(sessionV1.VisitorId)
				require.True(t, ok)
				require.Equal(t, storedSession, sessionV2)

				require.Equal(t, float64(1),
					testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
				require.Equal(t, float64(1),
					testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
						prometheus.Labels{"type": "inserted"}))
				require.Equal(t, float64(0),
					testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
						prometheus.Labels{"type": "expired"}))
				require.Equal(t, float64(0),
					testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
				require.Equal(t, float64(0),
					testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
			})

			t.Run("DifferentSessionUuid", func(t *testing.T) {
				promRegistry := prometheus.NewRegistry()
				service := ProvideService(logger, cfg, promRegistry)

				pageviews := uint16(rand.Uint32())

				sessionV1 := event.Session{
					VisitorId: "prisme_XXX",
					Pageviews: pageviews,
				}

				upserted := service.UpsertSession(sessionV1)
				require.True(t, upserted)

				sessionV2 := sessionV1
				sessionV2.SessionUuid = uuid.Must(uuid.NewV7())

				upserted = service.UpsertSession(sessionV2)
				require.True(t, upserted)

				// Upserting v1 should fail.
				upserted = service.UpsertSession(sessionV1)
				require.False(t, upserted)

				storedSession, ok := service.GetSession(sessionV1.VisitorId)
				require.True(t, ok)
				require.Equal(t, storedSession, sessionV2)

				require.Equal(t, float64(1),
					testutils.GaugeValue(t, promRegistry, "sessionstorage_active_sessions", nil))
				require.Equal(t, float64(2),
					testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
						prometheus.Labels{"type": "inserted"}))
				require.Equal(t, float64(0),
					testutils.CounterValue(t, promRegistry, "sessionstorage_sessions_total",
						prometheus.Labels{"type": "expired"}))
				require.Equal(t, float64(pageviews),
					testutils.HistogramSumValue(t, promRegistry, "sessionstorage_sessions_pageviews", nil))
				require.Equal(t, float64(0),
					testutils.CounterValue(t, promRegistry, "sessionstorage_get_session_misses", nil))
			})
		})
	})
}
