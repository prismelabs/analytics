package handlers

import "github.com/gofiber/fiber/v2"

type NotFound fiber.Handler

func ProvideNotFound() NotFound {
	return func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	}
}
