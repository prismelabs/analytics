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
	"github.com/tidwall/gjson"
)

type PostEventsCustom fiber.Handler

// ProvidePostEventsCustom is a wire provider for POST /api/v1/events/custom/:name events handler.
func ProvidePostEventsCustom(
	logger zerolog.Logger,
	eventStore eventstore.Service,
	uaParserService uaparser.Service,
	ipGeolocatorService ipgeolocator.Service,
	saltManagerService saltmanager.Service,
	storage storage.Storage,
) PostEventsCustom {
	return func(c *fiber.Ctx) error {
		if utils.UnsafeString(c.Request().Header.ContentType()) != fiber.MIMEApplicationJSON {
			return fiber.NewError(fiber.StatusBadRequest, "content type is not application/json")
		}

		customEv := event.Custom{}

		// Parse user agent.
		userAgent := utils.CopyBytes(c.Request().Header.UserAgent())
		customEv.Client = uaParserService.ParseUserAgent(utils.UnsafeString(userAgent))
		if customEv.Client.IsBot {
			return nil
		}

		// Event date and name.
		customEv.Timestamp = time.Now().UTC()
		customEv.Name = utils.CopyString(c.Params("name"))

		// Referrer of the POST request, that is the viewed page.
		requestReferrer := peekReferrerHeader(c)
		err := customEv.PageUri.Parse(requestReferrer)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Referer or X-Prisme-Referrer")
		}

		err = customEv.ReferrerUri.Parse(c.Request().Header.Peek("X-Prisme-Document-Referrer"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid X-Prisme-Document-Referrer")
		}

		// Find country code for given IP.
		customEv.CountryCode = ipGeolocatorService.FindCountryCodeForIP(c.IP())

		// Validate properties.
		body := utils.CopyBytes(c.Body())
		if len(body) > 0 {
			result := gjson.GetManyBytes(utils.CopyBytes(c.Body()), "@keys", "@values")
			result[0].ForEach(func(_, key gjson.Result) bool {
				customEv.Keys = append(customEv.Keys, key.String())
				return true
			})
			result[1].ForEach(func(_, value gjson.Result) bool {
				customEv.Values = append(customEv.Values, value.Raw)
				return true
			})
		}

		// Compute visitor ID.
		customEv.VisitorId = computeVisitorId(
			userAgent, saltManagerService.DailySalt().Bytes(), []byte(c.IP()),
			customEv.PageUri.Host(),
		)

		// Retrieve session ID.
		sessionBytes, err := storage.Get(sessionKey(customEv.VisitorId))
		if err != nil {
			logger.Err(err).Msg("failed to retrieve session id")
			return fiber.NewError(fiber.StatusInternalServerError, "failed to retrieve session id")
		}
		if sessionBytes == nil {
			return fiber.NewError(fiber.StatusBadRequest, "entry page missing")
		}
		session := unsafeBytesToSessionCast(sessionBytes)
		customEv.SessionId = session.id

		// Store event.
		err = eventStore.StoreCustom(c.UserContext(), &customEv)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to store custom event")
		}

		return nil
	}
}
