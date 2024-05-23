package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/storage"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
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
	storage storage.Storage,
) PostEventsPageview {
	return func(c *fiber.Ctx) error {
		// Referrer of the POST request, that is the viewed page.
		requestReferrer := peekReferrerHeader(c)

		pageView := event.PageView{}

		// Parse user agent.
		userAgent := utils.CopyBytes(c.Request().Header.UserAgent())
		pageView.Client = uaParserService.ParseUserAgent(utils.UnsafeString(userAgent))
		if pageView.Client.IsBot {
			return nil
		}

		// Event date.
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
		pageView.CountryCode = ipGeolocatorService.FindCountryCodeForIP(c.IP())

		// Compute visitor id.
		pageView.VisitorId = computeVisitorId(
			userAgent, saltManagerService.DailySalt().Bytes(), []byte(c.IP()),
			pageView.PageUri.Host(),
		)

		newSession := !equalBytes(pageView.ReferrerUri.Host(), pageView.PageUri.Host())
		var session session
		if newSession {
			// Compute session ID.
			session.id = computeSessionId(&pageView)
			session.entryTime = pageView.Timestamp

			// Store it.
			err := storage.Set(
				sessionKey(pageView.VisitorId),
				unsafeSessionToBytesCast(&session),
				24*time.Hour,
			)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
		} else {
			// Retrieve session.
			sessionBytes, err := storage.Get(sessionKey(pageView.VisitorId))
			if err != nil {
				logger.Err(err).Msg("failed to retrieve session id")
				return fiber.NewError(fiber.StatusInternalServerError, "failed to retrieve session id")
			}
			if sessionBytes == nil {
				return fiber.NewError(fiber.StatusBadRequest, "entry page missing")
			}

			session = *unsafeBytesToSessionCast(sessionBytes)
		}

		// Add session related fields.
		pageView.SessionId = session.id
		pageView.EntryTimestamp = session.entryTime

		err = eventStore.StorePageView(c.UserContext(), &pageView)
		if err != nil {
			logger.Err(err).Msg("failed to store page view event")
			return fiber.NewError(fiber.StatusInternalServerError, "failed to store page view event")
		}

		return nil
	}
}
