package middlewares

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
	"github.com/stretchr/testify/require"
)

func TestAccessLog(t *testing.T) {
	t.Run("WithoutRequestIdMiddleware/Panics", func(t *testing.T) {
		accessLogger := log.NewLogger("access_log", io.Discard, false)

		e := echo.New()
		h := AccessLog(accessLogger)(func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)

		require.Panics(t, func() {
			err := h(c)
			require.NoError(t, err)
		})
	})

	t.Run("WithRequestIdMiddleware", func(t *testing.T) {
		accessLoggerOutput := bytes.Buffer{}
		accessLogger := log.NewLogger("access_log", &accessLoggerOutput, false)

		e := echo.New()
		h := RequestId(config.Server{
			TrustProxy: true,
		})(AccessLog(accessLogger)(func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		}))

		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		res := httptest.NewRecorder()
		c := e.NewContext(req, res)

		err := h(c)
		require.NoError(t, err)

		actual := accessLoggerOutput.String()
		require.Regexp(t,
			`{"v":0,`+
				`"pid":\d+,`+
				`"hostname":"[^"]+",`+
				`"name":"access_log",`+
				`"request_id":"[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}",`+
				`"duration_ms":\d+(.\d+)?,`+
				`"source_ip":"`+extractIP(req)+`",`+
				`"method":"GET",`+
				`"path":"/hello",`+
				`"status_code":200,`+
				`"level":30,`+
				`"time":"((?:(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}(?:\.\d+)?))(Z|[\+-]\d{2}:\d{2})?)",`+
				`"msg":"request handled"}`,
			actual,
		)

		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, "Hello, World!", res.Body.String())
	})
}

func extractIP(req *http.Request) string {
	ra, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ra
}
