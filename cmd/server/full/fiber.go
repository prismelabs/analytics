package full

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/handlers"
	"github.com/prismelabs/analytics/pkg/middlewares"
	"github.com/prismelabs/analytics/pkg/wired"
)

// ProvideFiber is a wire provider for fiber.App.
func ProvideFiber(
	eventsCorsMiddleware middlewares.EventsCors,
	eventsRateLimiterMiddleware middlewares.EventsRateLimiter,
	minimalFiber wired.MinimalFiber,
	nonRegisteredOriginFilterMiddleware middlewares.NonRegisteredOriginFilter,
	noscriptHandlersCacheMiddleware middlewares.NoscriptHandlersCache,
	getNoscriptCustomEventHandler handlers.GetNoscriptEventsCustom,
	getNoscriptPageViewEventHandler handlers.GetNoscriptEventsPageviews,
	postCustomEventHandler handlers.PostEventsCustom,
	postIdentifyEventHandler handlers.PostEventsIdentify,
	postPageViewEventHandler handlers.PostEventsPageview,
) *fiber.App {
	app := (*fiber.App)(minimalFiber)

	// Public endpoints.
	app.Use("/api/v1/events/*",
		fiber.Handler(eventsCorsMiddleware),
		fiber.Handler(eventsRateLimiterMiddleware),
		fiber.Handler(nonRegisteredOriginFilterMiddleware),
	)

	app.Use("/api/v1/noscript/events/*",
		fiber.Handler(eventsCorsMiddleware),
		fiber.Handler(eventsRateLimiterMiddleware),
		fiber.Handler(nonRegisteredOriginFilterMiddleware),
		// Prevent caching of GET responses.
		fiber.Handler(noscriptHandlersCacheMiddleware),
	)

	app.Post("/api/v1/events/pageviews", fiber.Handler(postPageViewEventHandler))
	app.Get("/api/v1/noscript/events/pageviews", fiber.Handler(getNoscriptPageViewEventHandler))

	app.Post("/api/v1/events/identify", fiber.Handler(postIdentifyEventHandler))

	app.Post("/api/v1/events/custom/:name", fiber.Handler(postCustomEventHandler))
	app.Get("/api/v1/noscript/events/custom/:name", fiber.Handler(getNoscriptCustomEventHandler))

	return app
}
