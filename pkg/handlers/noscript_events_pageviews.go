package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/embedded"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstorage"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/rs/zerolog"
)

type GetNoscriptEventsPageview fiber.Handler

// ProvideGetNoscriptEventsPageview is a wire provider for
// GET /api/v1/noscript/events/pageview handler.
func ProvideGetNoscriptEventsPageview(
	logger zerolog.Logger,
	eventStore eventstore.Service,
	uaParserService uaparser.Service,
	ipGeolocatorService ipgeolocator.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
) GetNoscriptEventsPageview {
	return func(c *fiber.Ctx) error {
		// Referrer of the POST request, that is the viewed page.
		requestReferrer := peekReferrerQueryOrHeader(c)

		err := eventsPageviewsHandler(
			c.UserContext(),
			logger,
			eventStore,
			uaParserService,
			ipGeolocatorService,
			saltManagerService,
			sessionStorage,
			c.Context().UserAgent(),
			utils.UnsafeBytes(c.Query("document-referrer")),
			requestReferrer,
			utils.UnsafeBytes(c.IP()),
			c.Query("visitor-id"),
		)
		if err != nil {
			return err
		}

		return c.Send(embedded.NoscriptGif)
	}
}
