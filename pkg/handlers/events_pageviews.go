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

		// Compute visitor id.
		visitorId := requestVisitorId(c.Request())
		if visitorId == "" {
			visitorId = computeVisitorId("prisme_",
				c.Request().Header.UserAgent(), saltManagerService.DailySalt().Bytes(),
				utils.UnsafeBytes(c.IP()), pageView.PageUri.Host(),
			)
		}

		isExternalReferrer := equalBytes(referrerUri.Host(), pageView.PageUri.Host())
		newSession := !isExternalReferrer

		// Retrieve session.
		if !newSession {
			var ok bool
			pageView.Session, ok = sessionStorage.GetSession(visitorId)

			// Session not found.
			// This can happen if tracking script is not installed on all pages,
			// visitor/session was upgraded (X-Prisme-Visitor-Id was added session
			// after some pageview) to an authenticated one or prisme instance was
			// restarted.
			if !ok {
				newSession = true
			} else {
				pageView.Session.Pageviews++
				pageView.Timestamp = time.Now().UTC()
			}
		}

		// Create session.
		if newSession {
			// Filter bot.
			client := uaParserService.ParseUserAgent(
				utils.UnsafeString(c.Request().Header.UserAgent()),
			)
			if client.IsBot {
				return fiber.NewError(fiber.StatusBadRequest, "bot session filtered")
			}

			sessionUuid, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("failed to generate session uuid: %w", err)
			}

			pageView.Session = event.Session{
				PageUri:     &pageView.PageUri,
				ReferrerUri: &referrerUri,
				Client:      client,
				CountryCode: ipGeolocatorService.FindCountryCodeForIP(c.IP()),
				VisitorId:   visitorId,
				SessionUuid: sessionUuid,
				Utm:         extractUtmParams(pageView.PageUri.QueryArgs()),
				Pageviews:   1,
			}
			pageView.Timestamp = pageView.Session.SessionTime()
		}

		// Update session in storage.
		upserted := sessionStorage.UpsertSession(pageView.Session)
		if !upserted {
			logger.Debug().Msg("session upsert was ignored, pageview ignored")
			return nil
		}

		// Store event.
		err = eventStore.StorePageView(c.UserContext(), &pageView)
		if err != nil {
			return fmt.Errorf("failed to store pageview event: %w", err)
		}

		return nil
	}
}
