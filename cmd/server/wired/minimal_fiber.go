package wired

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/handlers"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
)

type MinimalFiber *fiber.App

// ProvideMinimalFiber is a wire provider for a minimally configured fiber.App
// with no route.
func ProvideMinimalFiber(
	cfg config.Server,
	viewsEngine fiber.Views,
	loggerMiddleware middlewares.Logger,
	accessLogMiddleware middlewares.AccessLog,
	requestIdMiddleware middlewares.RequestId,
	staticMiddleware middlewares.Static,
	healthcheckHandler handlers.HealhCheck,
) MinimalFiber {
	fiberCfg := fiber.Config{
		ServerHeader:          "prisme",
		StrictRouting:         true,
		AppName:               "Prisme Analytics",
		DisableStartupMessage: true,
		ErrorHandler: func(_ *fiber.Ctx, _ error) error {
			// Errors are handled manually by a middleware.
			return nil
		},
		Views:       viewsEngine,
		ViewsLayout: "layouts/empty",
	}
	if cfg.TrustProxy {
		fiberCfg.EnableIPValidation = false
		fiberCfg.ProxyHeader = cfg.ProxyHeader
	} else {
		fiberCfg.EnableIPValidation = true
		fiberCfg.ProxyHeader = ""
	}

	app := fiber.New(fiberCfg)

	app.Use(fiber.Handler(requestIdMiddleware))
	app.Use(fiber.Handler(accessLogMiddleware))
	app.Use(fiber.Handler(loggerMiddleware))

	app.Use("/static", fiber.Handler(staticMiddleware))

	app.Use("/api/v1/healthcheck", fiber.Handler(healthcheckHandler))

	return app
}
