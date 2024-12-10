package wired

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/handlers"
	"github.com/prismelabs/analytics/pkg/middlewares"
)

// ProvideFiber is a wire provider for fiber.App.
func ProvideFiber(
	apiEventsTimeoutMiddleware middlewares.ApiEventsTimeout,
	eventsCorsMiddleware middlewares.EventsCors,
	eventsRateLimiterMiddleware middlewares.EventsRateLimiter,
	getNoscriptCustomEventHandler handlers.GetNoscriptEventsCustom,
	getNoscriptOutboundLinksEventHandler handlers.GetNoscriptEventsOutboundLinks,
	getNoscriptPageViewsEventHandler handlers.GetNoscriptEventsPageviews,
	minimalFiber MinimalFiber,
	nonRegisteredOriginFilterMiddleware middlewares.NonRegisteredOriginFilter,
	noscriptHandlersCacheMiddleware middlewares.NoscriptHandlersCache,
	postCustomEventHandler handlers.PostEventsCustom,
	postFileDownloadsEventHandler handlers.PostEventsFileDownloads,
	postOutboundLinksEventHandler handlers.PostEventsOutboundLinks,
	postPageViewsEventHandler handlers.PostEventsPageviews,
	referrerAsDefaultOriginMiddleware middlewares.ReferrerAsDefaultOrigin,
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
		fiber.Handler(referrerAsDefaultOriginMiddleware),
		fiber.Handler(nonRegisteredOriginFilterMiddleware),
		fiber.Handler(apiEventsTimeoutMiddleware),
		// Prevent caching of GET responses.
		fiber.Handler(noscriptHandlersCacheMiddleware),
	)

	app.Post("/api/v1/events/pageviews", fiber.Handler(postPageViewsEventHandler))
	app.Get("/api/v1/noscript/events/pageviews", fiber.Handler(getNoscriptPageViewsEventHandler))

	app.Post("/api/v1/events/custom/:name", fiber.Handler(postCustomEventHandler))
	app.Get("/api/v1/noscript/events/custom/:name", fiber.Handler(getNoscriptCustomEventHandler))

	app.Post("/api/v1/events/outbound-links", fiber.Handler(postOutboundLinksEventHandler))
	app.Get("/api/v1/noscript/events/outbound-links", fiber.Handler(getNoscriptOutboundLinksEventHandler))

	app.Post("/api/v1/events/file-downloads", fiber.Handler(postFileDownloadsEventHandler))

	return app
}
