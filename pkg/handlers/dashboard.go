package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/prismelabs/analytics/pkg/embedded"
)

// Dashboad returns a static HTTP handler for /dashboard requests.
func Dashboard() fiber.Handler {
	fsCfg := filesystem.Config{
		Root:       http.FS(embedded.Dashboard),
		PathPrefix: "dashboard",
		Browse:     false,
		MaxAge:     3600,
	}

	return filesystem.New(fsCfg)
}
