package sessionstorage

import (
	"io"
	"testing"
	"time"

	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/log"
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
			service := ProvideService(logger, cfg)

			storedSession, ok := service.GetSession("...")
			require.False(t, ok)
			require.Equal(t, storedSession, event.Session{})
		})

		t.Run("Expired", func(t *testing.T) {
			service := ProvideService(logger, Config{
				gcInterval:         30 * time.Millisecond,
				sessionInactiveTtl: 30 * time.Millisecond,
			})

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
		})

		t.Run("Deleted", func(t *testing.T) {
			service := ProvideService(logger, cfg)

			session := event.Session{
				VisitorId: "prisme_XXX",
				Pageviews: 1,
			}

			upserted := service.UpsertSession(session)
			require.True(t, upserted)

			service.DeleteSession(session.VisitorId)

			storedSession, ok := service.GetSession(session.VisitorId)
			require.False(t, ok)
			require.Equal(t, storedSession, event.Session{})
		})
	})

	t.Run("UpsertSession", func(t *testing.T) {
		t.Run("NonExistent", func(t *testing.T) {
			service := ProvideService(logger, cfg)

			session := event.Session{
				VisitorId: "prisme_XXX",
				Pageviews: 1,
			}

			upserted := service.UpsertSession(session)
			require.True(t, upserted)

			storedSession, ok := service.GetSession(session.VisitorId)
			require.True(t, ok)
			require.Equal(t, storedSession, session)
		})

		t.Run("Existent", func(t *testing.T) {
			t.Run("NewerVersion", func(t *testing.T) {
				service := ProvideService(logger, cfg)

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

				storedSession, ok := service.GetSession(sessionV1.VisitorId)
				require.True(t, ok)
				require.Equal(t, storedSession, sessionV2) // session v2 overwrite session v1
			})

			t.Run("OlderVersion", func(t *testing.T) {
				service := ProvideService(logger, cfg)

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

				// Older version
				service.UpsertSession(sessionV1)

				storedSession, ok := service.GetSession(sessionV1.VisitorId)
				require.True(t, ok)
				require.Equal(t, storedSession, sessionV2)
			})
		})
	})
}
