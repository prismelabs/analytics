package handlers

import "github.com/gofiber/fiber/v2"

// HealthCheck returns a GET healthcheck handler.
func HealthCheck() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Response().SetStatusCode(fiber.StatusOK)
		return nil
	}
}
