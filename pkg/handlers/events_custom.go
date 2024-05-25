package handlers

import (
	"math/big"
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

		// Referrer of the POST request, that is the viewed page.
		err := customEv.PageUri.Parse(peekReferrerHeader(c))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Referer or X-Prisme-Referrer")
		}

		// Compute visitor ID.
		customEv.VisitorId = computeVisitorId(
			c.Request().Header.UserAgent(), saltManagerService.DailySalt().Bytes(),
			utils.UnsafeBytes(c.IP()), customEv.PageUri.Host(),
		)

		// Retrieve session.
		sessionIdBytes, err := storage.Get(sessionKey(customEv.VisitorId))
		if err != nil {
			logger.Err(err).Msg("failed to retrieve session id")
			return fiber.NewError(fiber.StatusInternalServerError, "failed to retrieve session id")
		}
		if sessionIdBytes == nil {
			return fiber.NewError(fiber.StatusBadRequest, "session missing")
		}
		customEv.SessionId = big.NewInt(0).SetBytes(sessionIdBytes)

		// Event date and name.
		customEv.Timestamp = time.Now().UTC()
		customEv.Name = utils.CopyString(c.Params("name"))

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

		// Store event.
		err = eventStore.StoreCustom(c.UserContext(), &customEv)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to store custom event")
		}

		return nil
	}
}
