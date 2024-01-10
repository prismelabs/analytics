package resterror

import (
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestFiberErrorHandler(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: FiberErrorHandler,
	})

	t.Run("FiberErrorOnly", func(t *testing.T) {
		app.Get("/fiber_error", func(c *fiber.Ctx) error {
			return fiber.NewError(fiber.StatusUnauthorized, "authorization header missing or invalid")
		})

		req := httptest.NewRequest("GET", "http://server.local/fiber_error", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, `{"error_code":"UnexpectedError","error_description":"authorization header missing or invalid"}`, string(body))
		require.Equal(t, resp.StatusCode, fiber.StatusUnauthorized)
	})

	t.Run("RestErrorOnly", func(t *testing.T) {
		app.Get("/rest_error", func(c *fiber.Ctx) error {
			return New("InvalidAuthorizationHeader")
		})

		req := httptest.NewRequest("GET", "http://server.local/rest_error", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, `{"error_code":"InvalidAuthorizationHeader","error_description":""}`, string(body))
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("FiberRestErrorsJoined", func(t *testing.T) {
		app.Get("/fiber_rest_error", func(c *fiber.Ctx) error {
			return New(
				"InvalidAuthorizationHeader",
				fiber.NewError(fiber.StatusUnauthorized, "invalid or missing authorization header"),
			)
		})

		req := httptest.NewRequest("GET", "http://server.local/fiber_rest_error", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, `{"error_code":"InvalidAuthorizationHeader","error_description":"invalid or missing authorization header"}`, string(body))
		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("GoError", func(t *testing.T) {
		app.Get("/go_error", func(c *fiber.Ctx) error {
			return errors.New("runtime error")
		})

		req := httptest.NewRequest("GET", "http://server.local/go_error", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, `{"error_code":"InternalServerError","error_description":"internal server error, check server logs for more information"}`, string(body))
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	})
}
