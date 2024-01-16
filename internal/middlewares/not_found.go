package middlewares

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

type NotFound fiber.Handler

// ProvideNotFound is a wire provider for 404 not found middleware.
func ProvideNotFound() NotFound {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil && errors.Is(err, fiber.ErrNotFound) {
			c.Context().SetStatusCode(fiber.StatusNotFound)
			return c.Render("not_found", fiber.Map{})
		}

		return err
	}
}
