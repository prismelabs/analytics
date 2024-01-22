package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/handlers"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
)

// ProvideFiber is a wire provider for fiber.App.
func ProvideFiber(
	cfg config.Config,
	viewsEngine fiber.Views,
	loggerMiddleware middlewares.Logger,
	accessLogMiddleware middlewares.AccessLog,
	requestIdMiddleware middlewares.RequestId,
	staticMiddleware middlewares.Static,
	withSessionMiddleware middlewares.WithSession,
	eventsCorsMiddleware middlewares.EventsCors,
	faviconMiddleware middlewares.Favicon,
	getSignUpHandler handlers.GetSignUp,
	postSignUpHander handlers.PostSignUp,
	getSignInHandler handlers.GetSignIn,
	postSignInHander handlers.PostSignIn,
	getIndexHander handlers.GetIndex,
	notFoundHandler handlers.NotFound,
	postPageViewEventHandler handlers.PostPageViewEvent,
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
		Views:       viewsEngine,
		ViewsLayout: "layouts/empty",
	}
	if cfg.Server.TrustProxy {
		fiberCfg.EnableIPValidation = false
		fiberCfg.ProxyHeader = fiber.HeaderXForwardedFor
	} else {
		fiberCfg.EnableIPValidation = true
		fiberCfg.ProxyHeader = ""
	}

	app := fiber.New(fiberCfg)

	app.Use(fiber.Handler(requestIdMiddleware))
	app.Use(fiber.Handler(accessLogMiddleware))
	app.Use(fiber.Handler(loggerMiddleware))

	// Public endpoints.
	app.Use(fiber.Handler(faviconMiddleware))

	app.Use("/api/v1/events/*", fiber.Handler(eventsCorsMiddleware))
	app.Post("/api/v1/events/pageviews", fiber.Handler(postPageViewEventHandler))

	app.Use("/static", fiber.Handler(staticMiddleware))

	app.Get("/sign_up", fiber.Handler(getSignUpHandler))
	app.Post("/sign_up", fiber.Handler(postSignUpHander))

	app.Get("/sign_in", fiber.Handler(getSignInHandler))
	app.Post("/sign_in", fiber.Handler(postSignInHander))

	// Authenticated endpoints.
	app.Use(fiber.Handler(withSessionMiddleware))

	app.Get("/", fiber.Handler(getIndexHander))

	// 404 not found handler.
	app.Use(fiber.Handler(notFoundHandler))

	return app
}
