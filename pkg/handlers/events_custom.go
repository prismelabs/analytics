package handlers

import (
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
)

type PostEventsCustom fiber.Handler

// ProvidePostEventsCustom is a wire provider for POST /api/v1/events/custom/:name events handler.
func ProvidePostEventsCustom(
	eventStore eventstore.Service,
	uaParserService uaparser.Service,
	ipgeolocatorService ipgeolocator.Service,
) PostEventsCustom {
	return func(c *fiber.Ctx) error {
		if utils.UnsafeString(c.Request().Header.ContentType()) != fiber.MIMEApplicationJSON {
			return fiber.NewError(fiber.StatusBadRequest, "content type is not application/json")
		}

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

		customEvent, err := event.NewCustom(domainName, c.Params("name"), c.Body())
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		// Store event.
		err = eventStore.StoreCustomEvent(c.UserContext(), customEvent)
		if err != nil {
			c.Response().SetStatusCode(fiber.StatusInternalServerError)
			return fmt.Errorf("failed to store pageview event: %w", err)
		}

		return nil
	}
}
