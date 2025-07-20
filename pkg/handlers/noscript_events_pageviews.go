package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/embedded"
	hutils "github.com/prismelabs/analytics/pkg/handlers/utils"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstore"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
)

// GetNoscriptEventsPageview returns a GET /api/v1/noscript/events/pageview
// handler.
func GetNoscriptEventsPageviews(
	eventStore eventstore.Service,
	uaParserService uaparser.Service,
	ipGeolocatorService ipgeolocator.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstore.Service,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Send(embedded.NoscriptGif)
		if err != nil {
			return err
		}

		// Referrer of the GET request, that is the viewed page.
		requestReferrer, err := hutils.PeekAndParseReferrerQueryHeader(c)
		if err != nil {
			return err
		}

		return eventsPageviewsHandler(
			c.UserContext(),
			eventStore,
			uaParserService,
			ipGeolocatorService,
			saltManagerService,
			sessionStorage,
			&c.Request().Header,
			requestReferrer,
			utils.UnsafeBytes(c.Query("document-referrer")),
			c.Context().UserAgent(),
			utils.UnsafeBytes(c.IP()),
			c.Query("status"),
			c.Query("visitor-id"),
		)
	}
}
