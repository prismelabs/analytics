package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// Cors middleware for /events handlers.
type EventsCors fiber.Handler

// ProvideEventsCors is a wire provider for /events cors middleware.
func ProvideEventsCors() EventsCors {
	return cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "POST",
	})
}
