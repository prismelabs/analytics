package sessions

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/prismelabs/prismeanalytics/internal/services/users"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService(t *testing.T) {
	t.Run("CreateSession", func(t *testing.T) {
		t.Run("StorageSetError", func(t *testing.T) {
			storageSetError := errors.New("failed to set value in storage")

			// Setup storage mock.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storage := NewMockStorage(ctrl)

			// Create store.
			sessionCfg := sessionConfig
			sessionCfg.Storage = storage
			sessionCfg.KeyGenerator = func() string { return "foo" }
			store := session.New(sessionCfg)

			// Create service.
			service := newService(store)

			// Setup sample fiber application.
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				err := service.CreateSession(c, users.NewUserId())
				require.Error(t, err)
				require.ErrorIs(t, err, storageSetError)

				return nil
			})

			// Setup mock to return an error on session creation.
			storage.EXPECT().Set(
				"foo",        // Session ID
				gomock.Any(), // Session data
				24*time.Hour).
				Return(storageSetError).
				Times(1)

			// Call handler.
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			_, err := app.Test(req)
			require.NoError(t, err)
		})

		t.Run("NoError", func(t *testing.T) {
			// Create store.
			store := session.New(sessionConfig)

			// Create service.
			service := newService(store)

			// Setup sample fiber application.
			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				err := service.CreateSession(c, users.NewUserId())
				require.NoError(t, err)

				return nil
			})

			// Call handler.
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			require.Len(t, resp.Cookies(), 1)
			sessCookie := resp.Cookies()[0]
			require.Equal(t, "prisme_session_id", sessCookie.Name)
			require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", sessCookie.Value)
			require.Equal(t, "/", sessCookie.Path)
			require.Equal(t, "", sessCookie.Domain)
			require.True(t, sessCookie.HttpOnly)
			require.True(t, sessCookie.Secure)
			require.Equal(t, http.SameSiteStrictMode, sessCookie.SameSite)
			require.Equal(t, "", sessCookie.RawExpires)
			require.Equal(t, int((24 * time.Hour).Seconds()), sessCookie.MaxAge)
		})
	})

	t.Run("GetSession", func(t *testing.T) {
		t.Run("NonExistentSession", func(t *testing.T) {
			// Create store.
			store := session.New(sessionConfig)

			// Create service.
			service := newService(store)

			app := fiber.New()
			app.Use(func(c *fiber.Ctx) error {
				session, err := service.GetSession(c)
				require.Error(t, err)
				require.ErrorIs(t, err, ErrSessionNotFound)
				require.Nil(t, session)

				return nil
			})

			// Call handler.
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			_, err := app.Test(req)
			require.NoError(t, err)
		})

		t.Run("ExistentSession", func(t *testing.T) {
			t.Run("AnonymousSession", func(t *testing.T) {
				// Create store.
				store := session.New(sessionConfig)

				// Create service.
				service := newService(store)

				app := fiber.New()
				reqCount := 0
				app.Use(func(c *fiber.Ctx) error {
					reqCount++

					if reqCount == 1 {
						// Create an anonymous session.
						session, err := store.Get(c)
						require.NoError(t, err)
						err = session.Save()
						require.NoError(t, err)
					} else if reqCount == 2 {
						// Try to retrieve session.
						userSession, err := service.GetSession(c)
						require.Error(t, err)
						require.ErrorIs(t, err, errAnonymousSession)
						require.Nil(t, userSession)
					} else {
						panic("unimplemented")
					}

					return nil
				})

				// Create session.
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				resp, err := app.Test(req)
				require.NoError(t, err)
				require.Len(t, resp.Cookies(), 1)

				// Get session.
				req = httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(resp.Cookies()[0])
				resp, err = app.Test(req)
				require.NoError(t, err)
				require.Len(t, resp.Cookies(), 1)
				sessCookie := resp.Cookies()[0]
				require.Positive(t, time.Now().Sub(sessCookie.Expires))
				require.Equal(t, 0, sessCookie.MaxAge)
			})

			t.Run("UserSession", func(t *testing.T) {
				// Create store.
				store := session.New(sessionConfig)

				// Create service.
				service := newService(store)

				app := fiber.New()
				reqCount := 0
				app.Use(func(c *fiber.Ctx) error {
					reqCount++

					if reqCount == 1 {
						err := service.CreateSession(c, users.NewUserId())
						require.NoError(t, err)
					} else if reqCount == 2 {
						session, err := service.GetSession(c)
						require.NoError(t, err)
						require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", session.UserId())
					} else {
						panic("unimplemented")
					}

					return nil
				})

				// Create session.
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				resp, err := app.Test(req)
				require.NoError(t, err)

				// Get session.
				req = httptest.NewRequest(http.MethodGet, "/", nil)
				req.AddCookie(resp.Cookies()[0])
				_, err = app.Test(req)
				require.NoError(t, err)
			})
		})
	})
}
