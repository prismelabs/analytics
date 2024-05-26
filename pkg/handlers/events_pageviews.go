package handlers

import (
	"math/big"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/storage"
	"github.com/google/uuid"
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
		session := event.Session{}

		// Referrer of the POST request, that is the viewed page.
		err := session.PageUri.Parse(peekReferrerHeader(c))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Referer or X-Prisme-Referrer")
		}
		// Referrer of the viewed page / document.
		err = session.ReferrerUri.Parse(c.Request().Header.Peek("X-Prisme-Document-Referrer"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid X-Prisme-Document-Referrer")
		}

		userAgent := utils.CopyBytes(c.Request().Header.UserAgent())

		// Compute visitor ID.
		session.VisitorId = computeVisitorId(
			userAgent, saltManagerService.DailySalt().Bytes(),
			utils.UnsafeBytes(c.IP()), session.PageUri.Host(),
		)

		newSession := !equalBytes(session.ReferrerUri.Host(), session.PageUri.Host())
		if newSession {
			// Parse user agent.
			session.Client = uaParserService.ParseUserAgent(utils.UnsafeString(userAgent))
			if session.Client.IsBot {
				return fiber.NewError(fiber.StatusBadRequest, "bot detected")
			}

			// Find country code for given IP.
			session.CountryCode = ipGeolocatorService.FindCountryCodeForIP(c.IP())

			// Generate session ID.
			session.SessionUuid, err = uuid.NewV7()
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
			session.Timestamp = time.Unix(session.SessionUuid.Time().UnixTime()).UTC()
			session.PageView.SessionId = big.NewInt(0).SetBytes(session.SessionUuid[:])

			// Store session id in KV store.
			err := storage.Set(
				sessionKey(session.VisitorId),
				session.SessionUuid[:],
				24*time.Hour,
			)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}

			// Store session event.
			err = eventStore.StoreSession(c.UserContext(), &session)
			if err != nil {
				logger.Err(err).Msg("failed to store session event")
				return fiber.NewError(fiber.StatusInternalServerError, "failed to store session event")
			}

		} else { // Page view event.

			// Retrieve session.
			sessionIdBytes, err := storage.Get(sessionKey(session.VisitorId))
			if err != nil {
				logger.Err(err).Msg("failed to retrieve session id")
				return fiber.NewError(fiber.StatusInternalServerError, "failed to retrieve session id")
			}
			if sessionIdBytes == nil {
				return fiber.NewError(fiber.StatusBadRequest, "session missing")
			}

			session.PageView.SessionId = big.NewInt(0).SetBytes(sessionIdBytes)
			session.Timestamp = time.Now().UTC()

			err = eventStore.StorePageView(c.UserContext(), &session.PageView)
			if err != nil {
				logger.Err(err).Msg("failed to store page view event")
				return fiber.NewError(fiber.StatusInternalServerError, "failed to store page view event")
			}
		}

		return nil
	}
}
