package ingestion

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/handlers"
	"github.com/prismelabs/analytics/pkg/middlewares"
	"github.com/prismelabs/analytics/pkg/wired"
)

// ProvideFiber is a wire provider for fiber.App.
func ProvideFiber(
	apiEventsTimeoutMiddleware middlewares.ApiEventsTimeout,
	eventsCorsMiddleware middlewares.EventsCors,
	eventsRateLimiterMiddleware middlewares.EventsRateLimiter,
	getNoscriptCustomEventHandler handlers.GetNoscriptEventsCustom,
	getNoscriptPageViewEventHandler handlers.GetNoscriptEventsPageviews,
	minimalFiber wired.MinimalFiber,
	nonRegisteredOriginFilterMiddleware middlewares.NonRegisteredOriginFilter,
	noscriptHandlersCacheMiddleware middlewares.NoscriptHandlersCache,
	postCustomEventHandler handlers.PostEventsCustom,
	postOutboundLinkEventHandler handlers.PostEventsOutboundLink,
	postPageViewEventHandler handlers.PostEventsPageviews,
) *fiber.App {
	app := (*fiber.App)(minimalFiber)

	// Public endpoints.
	app.Use("/api/v1/events/*",
		fiber.Handler(eventsCorsMiddleware),
		fiber.Handler(eventsRateLimiterMiddleware),
		fiber.Handler(nonRegisteredOriginFilterMiddleware),
		fiber.Handler(apiEventsTimeoutMiddleware),
	)

	app.Use("/api/v1/noscript/events/*",
		fiber.Handler(eventsCorsMiddleware),
		fiber.Handler(eventsRateLimiterMiddleware),
		fiber.Handler(nonRegisteredOriginFilterMiddleware),
		fiber.Handler(apiEventsTimeoutMiddleware),
		// Prevent caching of GET responses.
		fiber.Handler(noscriptHandlersCacheMiddleware),
	)

	app.Post("/api/v1/events/pageviews", fiber.Handler(postPageViewEventHandler))
	app.Get("/api/v1/noscript/events/pageviews", fiber.Handler(getNoscriptPageViewEventHandler))

	app.Post("/api/v1/events/custom/:name", fiber.Handler(postCustomEventHandler))
	app.Get("/api/v1/noscript/events/custom/:name", fiber.Handler(getNoscriptCustomEventHandler))

	return app
}
