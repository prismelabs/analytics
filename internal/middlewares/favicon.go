package middlewares

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/prismelabs/prismeanalytics/internal/embedded"
)

type Favicon fiber.Handler

// ProvideFavicon is a wire provider for in memory favicon middleware.
func ProvideFavicon() Favicon {
	return favicon.New(favicon.Config{
		File:       "static/favicon.ico",
		URL:        "",
		FileSystem: http.FS(embedded.Static),
	})
}
