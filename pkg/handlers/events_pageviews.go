package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/rs/zerolog"
)

type PostEventsPageview fiber.Handler

// ProvidePostEventsPageViews is a wire provider for POST /api/v1/events/pageviews events handler.
func ProvidePostEventsPageViews(
	logger zerolog.Logger,
	eventStore eventstore.Service,
	uaParserService uaparser.Service,
	ipgeolocatorService ipgeolocator.Service,
) PostEventsPageview {
	return func(c *fiber.Ctx) error {
		// Referrer of the POST request, that is the viewed page.
		requestReferrer := peekReferrerHeader(c)

		pageView := event.PageView{}
		pageView.Timestamp = time.Now().UTC()

		err := pageView.PageUri.Parse(requestReferrer)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Referer or X-Prisme-Referrer")
		}

		err = pageView.ReferrerUri.Parse(c.Request().Header.Peek("X-Prisme-Document-Referrer"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid X-Prisme-Document-Referrer")
		}

		// Find country code for given IP.
		pageView.CountryCode = ipgeolocatorService.FindCountryCodeForIP(c.IP())

		// Parse user agent.
		pageView.Client = uaParserService.ParseUserAgent(string(c.Request().Header.UserAgent()))

		err = eventStore.StorePageView(c.UserContext(), &pageView)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to store page view event")
		}

		return nil
	}
}
