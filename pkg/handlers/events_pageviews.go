package handlers

import (
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
)

type PostEventsPageview fiber.Handler

// ProvidePostEventsPageViews is a wire provider for POST /api/v1/events/pageviews events handler.
func ProvidePostEventsPageViews(
	eventStore eventstore.Service,
	uaParserService uaparser.Service,
	ipgeolocatorService ipgeolocator.Service,
) PostEventsPageview {
	return func(c *fiber.Ctx) error {
		// Referrer of the POST request, that is the viewed page.
		pageReferrer := string(peekReferrerHeader(c))
		pageUrl, err := url.ParseRequestURI(pageReferrer)
		if err != nil {
			c.Response().SetStatusCode(fiber.StatusBadRequest)
			return fmt.Errorf("invalid referrer: %w", err)
		}

		// Parse domain name.
		domainName, err := event.ParseDomainName(pageUrl.Hostname())
		if err != nil {
			c.Response().SetStatusCode(fiber.StatusBadRequest)
			return fmt.Errorf("invalid referrer hostname: %w", err)
		}

		// Website from which viewer comes from.
		referrer := string(c.Request().Header.Peek("X-Prisme-Document-Referrer"))

		// Parse user agent.
		cli := uaParserService.ParseUserAgent(string(c.Request().Header.UserAgent()))

		// Find country code for given IP.
		countryCode := ipgeolocatorService.FindCountryCodeForIP(c.IP())

		// Create pageview.
		pageview, err := event.NewPageView(pageUrl, domainName, cli, referrer, countryCode)
		if err != nil {
			c.Response().SetStatusCode(fiber.StatusBadRequest)
			return fmt.Errorf("invalid pageview event: %w", err)
		}

		// Store event.
		err = eventStore.StorePageViewEvent(c.UserContext(), pageview)
		if err != nil {
			c.Response().SetStatusCode(fiber.StatusInternalServerError)
			return fmt.Errorf("failed to store pageview event: %w", err)
		}

		return nil
	}
}
