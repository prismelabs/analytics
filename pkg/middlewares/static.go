package middlewares

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/embedded"
)

type Static fiber.Handler

func ProvideStatic(cfg config.Server) Static {
	fsCfg := filesystem.Config{
		Root:       http.FS(embedded.Static),
		PathPrefix: "static",
		Browse:     false,
	}

	if cfg.Debug {
		fsCfg = filesystem.Config{
			Root:       http.Dir("internal/embedded/static"),
			PathPrefix: "",
			Browse:     true,
		}
	}

	return filesystem.New(fsCfg)
}
