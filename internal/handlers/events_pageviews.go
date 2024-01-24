package handlers

import (
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/event"
	"github.com/prismelabs/prismeanalytics/internal/services/eventstore"
	"github.com/prismelabs/prismeanalytics/internal/services/sourceregistry"
	"github.com/prismelabs/prismeanalytics/internal/services/uaparser"
)

type PostPageViewEvent fiber.Handler

// ProvidePostEventsPageViews is a wire provider for POST pageview events handler.
func ProvidePostEventsPageViews(
	eventStore eventstore.Service,
	sourceRegistry sourceregistry.Service,
	uaParserService uaparser.Service,
) PostPageViewEvent {
	return func(c *fiber.Ctx) error {
		// Referrer of the POST request, that is the viewed page.
		pageReferrer := string(peekReferrerHeader(c))
		pageUrl, err := url.ParseRequestURI(pageReferrer)
		if err != nil {
			c.Response().SetStatusCode(fiber.StatusBadRequest)
			return fmt.Errorf("invalid referrer: %w", err)
		}

		// Website from which viewer comes from.
		referrer := string(c.Request().Header.Peek("X-Prisme-Document-Referrer"))

		// Parse user agent.
		cli := uaParserService.ParseUserAgent(string(c.Request().Header.UserAgent()))

		// Create pageview.
		pageview, err := event.NewPageView(pageUrl, cli, referrer)
		if err != nil {
			c.Response().SetStatusCode(fiber.StatusBadRequest)
			return fmt.Errorf("invalid pageview event: %w", err)
		}

		// Ensure source is registered.
		isRegistered, err := sourceRegistry.IsSourceRegistered(c.UserContext(), pageview.DomainName)
		if err != nil {
			c.Response().SetStatusCode(fiber.StatusInternalServerError)
			return fmt.Errorf("failed to store pageview event: %w", err)
		}

		// Source is not registered.
		if !isRegistered {
			c.Response().SetStatusCode(fiber.StatusBadRequest)
			return fmt.Errorf("source %q not registered", pageview.DomainName.SourceString())
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
