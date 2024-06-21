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

		// Compute device id.
		deviceId := computeDeviceId(
			saltManagerService.StaticSalt().Bytes(), c.Request().Header.UserAgent(),
			utils.UnsafeBytes(c.IP()), pageView.PageUri.Host(),
		)

		// Peek visitor id.
		visitorId := string(c.Request().Header.Peek("X-Prisme-Visitor-Id"))

		isExternalReferrer := equalBytes(referrerUri.Host(), pageView.PageUri.Host())
		newSession := !isExternalReferrer

		// Retrieve session.
		if !newSession {
			sessionExists := false
			if visitorId != "" {
				_, sessionExists = sessionStorage.IdentifySession(deviceId, visitorId)
			} else {
				_, sessionExists = sessionStorage.GetSession(deviceId)
			}

			// Session not found.
			// This can happen if tracking script is not installed on all pages,
			// or prisme instance was restarted.
			if !sessionExists {
				newSession = true
			} else {
				pageView.Session, sessionExists = sessionStorage.IncSessionPageviewCount(deviceId)
				if !sessionExists {
					logger.Panic().Msg("failed to increment session pageview count after GetSession returned a session: session not found")
				}
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

			// Compute visitor id if none was provided along request.
			if visitorId == "" {
				visitorId = computeVisitorId("prisme_",
					saltManagerService.DailySalt().Bytes(), c.Request().Header.UserAgent(),
					utils.UnsafeBytes(c.IP()), pageView.PageUri.Host(), []byte(deviceId),
				)
			}

			pageView.Session = event.Session{
				PageUri:       &pageView.PageUri,
				ReferrerUri:   &referrerUri,
				Client:        client,
				CountryCode:   ipGeolocatorService.FindCountryCodeForIP(c.IP()),
				VisitorId:     visitorId,
				SessionUuid:   sessionUuid,
				Utm:           extractUtmParams(pageView.PageUri.QueryArgs()),
				PageviewCount: 1,
			}
			pageView.Timestamp = pageView.Session.SessionTime()

			sessionStorage.InsertSession(deviceId, pageView.Session)
		}

		// Store event.
		err = eventStore.StorePageView(c.UserContext(), &pageView)
		if err != nil {
			return fmt.Errorf("failed to store pageview event: %w", err)
		}

		return nil
	}
}
