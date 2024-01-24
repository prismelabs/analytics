package full

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/cmd/server/wired"
	"github.com/prismelabs/prismeanalytics/internal/handlers"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
)

// ProvideFiber is a wire provider for fiber.App.
func ProvideFiber(
	eventsCorsMiddleware middlewares.EventsCors,
	eventsRateLimiterMiddleware middlewares.EventsRateLimiter,
	faviconMiddleware middlewares.Favicon,
	getIndexHander handlers.GetIndex,
	getSignInHandler handlers.GetSignIn,
	getSignUpHandler handlers.GetSignUp,
	minimalFiber wired.MinimalFiber,
	notFoundHandler handlers.NotFound,
	postPageViewEventHandler handlers.PostPageViewEvent,
	postSignInHander handlers.PostSignIn,
	postSignUpHander handlers.PostSignUp,
	withSessionMiddleware middlewares.WithSession,
) *fiber.App {
	app := (*fiber.App)(minimalFiber)

	// Public endpoints.
	app.Use(fiber.Handler(faviconMiddleware))

	app.Use("/api/v1/events/*",
		fiber.Handler(eventsCorsMiddleware),
		fiber.Handler(eventsRateLimiterMiddleware),
	)
	app.Post("/api/v1/events/pageviews", fiber.Handler(postPageViewEventHandler))

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
