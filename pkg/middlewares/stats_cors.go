package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/prismelabs/analytics/pkg/prisme"
)

// StatsCors returns a cors middleware for /api/*/stats/* handlers.
func StatsCors(cfg prisme.Config) fiber.Handler {
	origins := cfg.ApiStatsAllowOrigins
	if origins == "" {
		origins = "http://localhost:5173"
	}

	return cors.New(cors.Config{
		AllowOrigins: origins,
		AllowMethods: "GET",
	})
}
