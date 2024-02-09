package handlers

import "github.com/gofiber/fiber/v2"

type HealhCheck fiber.Handler

// ProvideHealthCheck is a wire provider for GET healthcheck handler.
func ProvideHealthCheck() HealhCheck {
	return func(c *fiber.Ctx) error {
		c.Response().SetStatusCode(fiber.StatusOK)
		return nil
	}
}
