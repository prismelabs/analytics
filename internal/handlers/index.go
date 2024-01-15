package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type GetIndex fiber.Handler

func ProvideGetIndex() GetIndex {
	return func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"title": "Home - Prisme Analytics",
		})
	}
}
