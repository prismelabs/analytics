package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/stretchr/testify/require"
)

func TestRequestIdMiddleware(t *testing.T) {
	t.Run("DoNotTrustProxy", func(t *testing.T) {
		cfg := config.Server{
			TrustProxy: false,
		}

		t.Run("WithoutRequestIdHeader", func(t *testing.T) {
			e := echo.New()
			h := RequestId(cfg)(func(c echo.Context) error {
				requestId := c.Get(RequestIdKey).(string)
				require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", requestId)
				return nil
			})

			req := httptest.NewRequest(http.MethodGet, "/hello", nil)
			res := httptest.NewRecorder()
			c := e.NewContext(req, res)

			err := h(c)
			require.NoError(t, err)
		})

		t.Run("WithRequestIdHeader", func(t *testing.T) {
			reqRequestId := uuid.New()

			e := echo.New()
			h := RequestId(cfg)(func(c echo.Context) error {
				requestId := c.Get(RequestIdKey).(string)
				require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", requestId)
				require.NotEqual(t, reqRequestId.String(), requestId)
				return nil
			})

			req := httptest.NewRequest(http.MethodGet, "/hello", nil)

			// Add request id.
			req.Header.Add(echo.HeaderXRequestID, reqRequestId.String())

			res := httptest.NewRecorder()
			c := e.NewContext(req, res)

			err := h(c)
			require.NoError(t, err)
		})
	})

	t.Run("TrustProxy", func(t *testing.T) {
		cfg := config.Server{
			TrustProxy: true,
		}

		t.Run("WithoutRequestIdHeader", func(t *testing.T) {
			e := echo.New()
			h := RequestId(cfg)(func(c echo.Context) error {
				requestId := c.Get(RequestIdKey).(string)
				require.Regexp(t, "[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}", requestId)
				return nil
			})

			req := httptest.NewRequest(http.MethodGet, "/hello", nil)
			res := httptest.NewRecorder()
			c := e.NewContext(req, res)

			err := h(c)
			require.NoError(t, err)
		})

		t.Run("WithRequestIdHeader", func(t *testing.T) {
			expectedRequestId := uuid.New()

			e := echo.New()
			h := RequestId(cfg)(func(c echo.Context) error {
				require.Equal(t, expectedRequestId.String(), c.Get(RequestIdKey))
				return nil
			})

			req := httptest.NewRequest(http.MethodGet, "/hello", nil)

			// Add request id.
			req.Header.Add(echo.HeaderXRequestID, expectedRequestId.String())

			res := httptest.NewRecorder()
			c := e.NewContext(req, res)

			err := h(c)
			require.NoError(t, err)
		})
	})
}
