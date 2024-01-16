package handlers

import "github.com/gofiber/fiber/v2"

type NotFound fiber.Handler

func ProvideNotFound() NotFound {
	return func(c *fiber.Ctx) error {
		c.Context().SetStatusCode(fiber.StatusNotFound)
		return c.Render("not_found", fiber.Map{})
	}
}
