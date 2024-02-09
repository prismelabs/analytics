package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/pkg/config"
	"github.com/prismelabs/prismeanalytics/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestAccessLog(t *testing.T) {
	t.Run("WithoutRequestIdMiddleware/Panics", func(t *testing.T) {
		accessLogger := log.NewLogger("access_log", io.Discard, false)

		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			require.Panics(t, func() {
				err := c.Next()
				require.NoError(t, err)
			})
			return nil
		})
		app.Use(accessLog(accessLogger))
		app.Use(func(c *fiber.Ctx) error {
			return nil
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)

		_, err := app.Test(req)
		require.NoError(t, err)
	})

	t.Run("WithRequestIdMiddleware", func(t *testing.T) {
		type testCase struct {
			name           string
			proxyHeader    string
			xForwardedFor  string
			loggedSourceIp string
		}

		testCases := []testCase{
			{
				name:           "TrustXForwardedFor",
				proxyHeader:    fiber.HeaderXForwardedFor,
				xForwardedFor:  "10.1.2.3",
				loggedSourceIp: "10.1.2.3",
			},
			{
				name:           "DoNotTrustXForwardedFor",
				proxyHeader:    "",
				xForwardedFor:  "10.1.2.3",
				loggedSourceIp: "0.0.0.0",
			},
		}

		for _, tcase := range testCases {
			t.Run(tcase.name, func(t *testing.T) {
				accessLoggerOutput := bytes.Buffer{}
				accessLogger := log.NewLogger("access_log", &accessLoggerOutput, false)

				app := fiber.New(fiber.Config{
					ProxyHeader: tcase.proxyHeader,
				})
				app.Use(fiber.Handler(ProvideRequestId(config.Server{})))
				app.Use(accessLog(accessLogger))
				app.Use(func(c *fiber.Ctx) error {
					return nil
				})

				req := httptest.NewRequest(http.MethodGet, "/hello", nil)
				req.Header.Add(fiber.HeaderXForwardedFor, tcase.xForwardedFor)

				_, err := app.Test(req)
				require.NoError(t, err)

				actual := accessLoggerOutput.String()
				require.Regexp(t,
					`{"v":0,`+
						`"pid":\d+,`+
						`"hostname":"[^"]+",`+
						`"name":"access_log",`+
						`"request_id":"[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}",`+
						`"duration_ms":\d+(.\d+)?,`+
						`"source_ip":"`+tcase.loggedSourceIp+`",`+
						`"method":"GET",`+
						`"path":"/hello",`+
						`"status_code":200,`+
						`"level":30,`+
						`"time":"((?:(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}(?:\.\d+)?))(Z|[\+-]\d{2}:\d{2})?)",`+
						`"msg":"request handled"}`,
					actual,
				)
			})
		}
	})
}
