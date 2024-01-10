package main

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
	"github.com/prismelabs/prismeanalytics/internal/renderer"
)

// ProvideFiber is a wire provider for fiber.App.
func ProvideFiber(
	cfg config.Config,
	accessLogger AccessLogger,
	renderer renderer.Renderer,
) *fiber.App {
	fiberCfg := fiber.Config{
		ServerHeader:          "prisme",
		StrictRouting:         true,
		AppName:               "Prisme Analytics",
		DisableStartupMessage: true,
		ErrorHandler: func(_ *fiber.Ctx, _ error) error {
			// Errors are handled manually by a middleware.
			return nil
		},
	}
	if cfg.Server.TrustProxy {
		fiberCfg.EnableIPValidation = false
		fiberCfg.ProxyHeader = fiber.HeaderXForwardedFor
	} else {
		fiberCfg.EnableIPValidation = true
		fiberCfg.ProxyHeader = ""
	}

	app := fiber.New(fiberCfg)

	app.Use(middlewares.RequestId(cfg.Server))
	app.Use(middlewares.AccessLog(accessLogger.Logger))

	// Error handler.
	// Handle error manually before access log middleware.
	app.Use(middlewares.RestError)

	return app
}
