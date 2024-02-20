package wired

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/handlers"
	"github.com/prismelabs/analytics/pkg/middlewares"
)

type MinimalFiber *fiber.App

// ProvideMinimalFiber is a wire provider for a minimally configured fiber.App
// with no route.
func ProvideMinimalFiber(
	accessLogMiddleware middlewares.AccessLog,
	errorHandlerMiddleware middlewares.ErrorHandler,
	fiberCfg fiber.Config,
	healthcheckHandler handlers.HealhCheck,
	loggerMiddleware middlewares.Logger,
	requestIdMiddleware middlewares.RequestId,
	staticMiddleware middlewares.Static,
) MinimalFiber {
	app := fiber.New(fiberCfg)

	app.Use(fiber.Handler(requestIdMiddleware))
	app.Use(fiber.Handler(accessLogMiddleware))
	app.Use(fiber.Handler(loggerMiddleware))
	app.Use(fiber.Handler(errorHandlerMiddleware))

	app.Use("/static", fiber.Handler(staticMiddleware))

	app.Use("/api/v1/healthcheck", fiber.Handler(healthcheckHandler))

	return app
}

// ProvideMinimalFiberConfig is a wire provider for fiber configuration.
func ProvideMinimalFiberConfig(
	cfg config.Server,
) fiber.Config {
	fiberCfg := fiber.Config{
		ServerHeader:          "prisme",
		StrictRouting:         true,
		AppName:               "Prisme Analytics",
		DisableStartupMessage: true,
		ErrorHandler: func(_ *fiber.Ctx, _ error) error {
			// Errors are handled by errorHandlerMiddleware so access log
			// contains right status code.
			return nil
		},
	}
	if cfg.TrustProxy {
		fiberCfg.EnableIPValidation = false
		fiberCfg.ProxyHeader = cfg.ProxyHeader
	} else {
		fiberCfg.EnableIPValidation = true
		fiberCfg.ProxyHeader = ""
	}

	return fiberCfg
}
