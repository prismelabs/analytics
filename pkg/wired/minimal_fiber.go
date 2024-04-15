package wired

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/handlers"
	"github.com/prismelabs/analytics/pkg/middlewares"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/rs/zerolog"
)

type MinimalFiber *fiber.App

// ProvideMinimalFiber is a wire provider for a minimally configured fiber.App
// with no route.
func ProvideMinimalFiber(
	accessLogMiddleware middlewares.AccessLog,
	errorHandlerMiddleware middlewares.ErrorHandler,
	fiberCfg fiber.Config,
	healthcheckHandler handlers.HealhCheck,
	logger zerolog.Logger,
	requestIdMiddleware middlewares.RequestId,
	nonRegistreredOriginFilterMiddleware middlewares.NonRegisteredOriginFilter,
	staticMiddleware middlewares.Static,
	teardownService teardown.Service,
) MinimalFiber {
	app := fiber.New(fiberCfg)

	teardownService.RegisterProcedure(func() error {
		logger.Info().Msg("shutting down fiber server...")
		err := app.Shutdown()
		logger.Err(err).Msg("fiber server shutdown.")

		return err
	})

	app.Use(fiber.Handler(requestIdMiddleware))
	app.Use(fiber.Handler(accessLogMiddleware))
	app.Use(fiber.Handler(errorHandlerMiddleware))

	app.Use("/static", fiber.Handler(nonRegistreredOriginFilterMiddleware),
		fiber.Handler(staticMiddleware))

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
