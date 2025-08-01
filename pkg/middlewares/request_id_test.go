package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/prisme"
	"github.com/stretchr/testify/require"
)

func TestRequestIdMiddleware(t *testing.T) {
	t.Run("DoNotTrustProxy", func(t *testing.T) {
		cfg := prisme.Config{
			TrustProxy: false,
		}

		t.Run("WithoutRequestIdHeader", func(t *testing.T) {
			middlewareCalled := false

			app := fiber.New()
			app.Use(RequestId(cfg))
			app.Use(func(c *fiber.Ctx) error {
				middlewareCalled = true

				requestId := c.Locals(RequestIdKey{}).(string)
				require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", requestId)
				return nil
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			res, err := app.Test(req)
			require.NoError(t, err)
			require.True(t, middlewareCalled)
			require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", res.Header.Get("X-Prisme-Request-Id"))
		})

		t.Run("WithRequestIdHeader", func(t *testing.T) {
			t.Run("Default", func(t *testing.T) {
				middlewareCalled := false
				reqRequestId := uuid.New()

				app := fiber.New()
				app.Use(RequestId(cfg))
				app.Use(func(c *fiber.Ctx) error {
					middlewareCalled = true

					requestId := c.Locals(RequestIdKey{}).(string)
					require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", requestId)
					require.NotEqual(t, reqRequestId.String(), requestId)

					return nil
				})

				req := httptest.NewRequest(http.MethodGet, "/", nil)
				// Add request id.
				req.Header.Add(fiber.HeaderXRequestID, reqRequestId.String())

				res, err := app.Test(req)
				require.NoError(t, err)
				require.True(t, middlewareCalled)
				require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", res.Header.Get("X-Prisme-Request-Id"))
				require.NotEqual(t, reqRequestId.String(), res.Header.Get("X-Prisme-Request-Id"))
			})

			t.Run("Custom", func(t *testing.T) {
				cfg := prisme.Config{
					TrustProxy:           false,
					ProxyRequestIdHeader: "X-Custom-Request-Id",
				}
				middlewareCalled := false

				app := fiber.New()
				app.Use(RequestId(cfg))
				app.Use(func(c *fiber.Ctx) error {
					middlewareCalled = true

					requestId := c.Locals(RequestIdKey{}).(string)
					require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", requestId)

					return nil
				})

				req := httptest.NewRequest(http.MethodGet, "/", nil)
				// Add request ids.
				req.Header.Add(fiber.HeaderXRequestID, "bar")
				req.Header.Add("X-Custom-Request-Id", "foo")

				res, err := app.Test(req)
				require.NoError(t, err)
				require.True(t, middlewareCalled)
				require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", res.Header.Get("X-Prisme-Request-Id"))
			})
		})
	})

	t.Run("TrustProxy", func(t *testing.T) {
		cfg := prisme.Config{
			TrustProxy:           true,
			ProxyRequestIdHeader: fiber.HeaderXRequestID,
		}

		t.Run("WithoutRequestIdHeader", func(t *testing.T) {
			middlewareCalled := false

			app := fiber.New()
			app.Use(RequestId(cfg))
			app.Use(func(c *fiber.Ctx) error {
				middlewareCalled = true

				requestId := c.Locals(RequestIdKey{}).(string)
				require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", requestId)
				return nil
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			res, err := app.Test(req)
			require.NoError(t, err)
			require.True(t, middlewareCalled)
			require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", res.Header.Get("X-Prisme-Request-Id"))
		})

		t.Run("WithRequestIdHeader", func(t *testing.T) {
			t.Run("Default", func(t *testing.T) {
				middlewareCalled := false
				expectedRequestId := uuid.New()

				app := fiber.New()
				app.Use(RequestId(cfg))
				app.Use(func(c *fiber.Ctx) error {
					middlewareCalled = true

					require.Equal(t, expectedRequestId.String(), c.Locals(RequestIdKey{}))
					return nil
				})
				app.Get("/", func(c *fiber.Ctx) error {
					return nil
				})

				req := httptest.NewRequest(http.MethodGet, "/", nil)
				// Add request id.
				req.Header.Add(fiber.HeaderXRequestID, expectedRequestId.String())

				res, err := app.Test(req)
				require.NoError(t, err)
				require.True(t, middlewareCalled)
				require.Equal(t, expectedRequestId.String(), res.Header.Get("X-Prisme-Request-Id"))
			})

			t.Run("Custom", func(t *testing.T) {
				cfg := prisme.Config{
					TrustProxy:           true,
					ProxyRequestIdHeader: "X-Custom-Request-Id",
				}

				middlewareCalled := false

				app := fiber.New()
				app.Use(RequestId(cfg))
				app.Use(func(c *fiber.Ctx) error {
					middlewareCalled = true

					require.Equal(t, "foo", c.Locals(RequestIdKey{}))
					return nil
				})
				app.Get("/", func(c *fiber.Ctx) error {
					return nil
				})

				req := httptest.NewRequest(http.MethodGet, "/", nil)
				// Add request id.
				req.Header.Add(fiber.HeaderXRequestID, "bar")
				req.Header.Add("X-Custom-Request-Id", "foo")

				res, err := app.Test(req)
				require.NoError(t, err)
				require.True(t, middlewareCalled)
				require.Equal(t, "foo", res.Header.Get("X-Prisme-Request-Id"))
			})
		})
	})
}
