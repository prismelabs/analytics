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
		MaxAge:     3600,
	}

	if cfg.Debug {
		fsCfg = filesystem.Config{
			Root:       http.Dir("pkg/embedded/static"),
			PathPrefix: "",
			Browse:     true,
		}
	}

	handler := filesystem.New(fsCfg)

	return func(c *fiber.Ctx) error {
		c.Response().Header.Add(fiber.HeaderAcceptCH, "Sec-CH-UA, Sec-CH-UA-Mobile, Sec-CH-UA-Platform")
		return handler(c)
	}
}
