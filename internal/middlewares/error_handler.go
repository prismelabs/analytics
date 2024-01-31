package middlewares

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

type ErrorHandler fiber.Handler

// ProvideErrorHandler is a wire provider for a simple error handler middleware.
func ProvideErrorHandler() ErrorHandler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		var fiberErr *fiber.Error
		if err != nil && errors.As(err, &fiberErr) {
			c.Response().SetStatusCode(fiberErr.Code)
		}

		return err
	}
}
