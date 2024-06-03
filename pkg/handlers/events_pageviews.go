package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstorage"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/rs/zerolog"
)

type PostEventsPageview fiber.Handler

// ProvidePostEventsPageViews is a wire provider for POST /api/v1/events/pageviews events handler.
func ProvidePostEventsPageViews(
	logger zerolog.Logger,
	eventStore eventstore.Service,
	uaParserService uaparser.Service,
	ipGeolocatorService ipgeolocator.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
) PostEventsPageview {
	return func(c *fiber.Ctx) error {
		// Referrer of the POST request, that is the viewed page.
		requestReferrer := peekReferrerHeader(c)

		referrerUri := event.ReferrerUri{}
		pageView := event.PageView{}

		// Parse referrer URI.
		err := referrerUri.Parse(utils.CopyBytes(c.Request().Header.Peek("X-Prisme-Document-Referrer")))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid X-Prisme-Document-Referrer")
		}

		// Parse page URI.
		err = pageView.PageUri.Parse(utils.CopyBytes(requestReferrer))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Referer or X-Prisme-Referrer")
		}

		userAgent := c.Request().Header.UserAgent()
		visitorId := computeVisitorId(
			userAgent, saltManagerService.DailySalt().Bytes(),
			utils.UnsafeBytes(c.IP()), pageView.PageUri.Host(),
		)

		newSession := !equalBytes(referrerUri.Host(), pageView.PageUri.Host())
		if newSession {
			sessionUuid, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("failed to generate session uuid: %w", err)
			}

			pageView.Session = event.Session{
				PageUri:     &pageView.PageUri,
				ReferrerUri: &referrerUri,
				Client:      uaParserService.ParseUserAgent(utils.UnsafeString(userAgent)),
				CountryCode: ipGeolocatorService.FindCountryCodeForIP(c.IP()),
				VisitorId:   visitorId,
				SessionUuid: sessionUuid,
				Utm:         extractUtmParams(pageView.PageUri.QueryArgs()),
				Pageviews:   1,
			}
			pageView.Timestamp = pageView.Session.SessionTime()

		} else { // session should already exists
			var ok bool
			pageView.Session, ok = sessionStorage.GetSession(visitorId)

			// Session not found.
			// This can happen if tracking script is not installed on all pages or
			// prisme instance was restarted.
			if !ok {
				return fiber.NewError(fiber.StatusBadRequest, "session not found")
			}

			pageView.Session.Pageviews++
			pageView.Timestamp = time.Now().UTC()
		}

		// Update session in storage.
		upserted := sessionStorage.UpsertSession(pageView.Session)
		if !upserted {
			logger.Debug().Msg("session upsert was ignored, pageview ignored")
			return nil
		}

		err = eventStore.StorePageView(c.UserContext(), &pageView)
		if err != nil {
			return fmt.Errorf("failed to store pageview event: %w", err)
		}

		return nil
	}
}
