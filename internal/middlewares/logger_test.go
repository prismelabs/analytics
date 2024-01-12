package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("WithoutRequestIdMiddleware/Panics", func(t *testing.T) {
		logger := log.NewLogger("app_log", io.Discard, false)

		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			require.Panics(t, func() {
				c.Next()
			})
			return nil
		})
		app.Use(fiber.Handler(ProvideLogger(logger)))
		app.Use(func(c *fiber.Ctx) error {
			return nil
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)

		_, err := app.Test(req)
		require.NoError(t, err)
	})

	t.Run("WithRequestIdMiddleware", func(t *testing.T) {
		loggerOutput := bytes.Buffer{}
		logger := log.NewLogger("app_log", &loggerOutput, false)

		app := fiber.New()
		app.Use(fiber.Handler(ProvideRequestId(config.Server{})))
		app.Use(fiber.Handler(ProvideLogger(logger)))
		app.Use(func(c *fiber.Ctx) error {
			logger := c.Locals(LoggerKey{}).(log.Logger)
			logger.Info().Msg("hello from middleware")
			return nil
		})

		req := httptest.NewRequest(http.MethodGet, "/hello", nil)

		_, err := app.Test(req)
		require.NoError(t, err)

		actual := loggerOutput.String()
		require.Regexp(t, `{"v":0,`+
			`"pid":\d+,`+
			`"hostname":"[^"]+",`+
			`"name":"app_log",`+
			`"request_id":"[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}",`+
			`"level":30,`+
			`"time":"((?:(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}(?:\.\d+)?))(Z|[\+-]\d{2}:\d{2})?)",`+
			`"msg":"hello from middleware"}`,
			actual)
	})
}
