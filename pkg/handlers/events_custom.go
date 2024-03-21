package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/tidwall/gjson"
)

type PostEventsCustoms fiber.Handler

// ProvidePostEventsCustoms is a wire provider for POST /api/v1/events/customs/:name events handler.
func ProvidePostEventsCustoms(
	eventStore eventstore.Service,
	uaParserService uaparser.Service,
	ipgeolocatorService ipgeolocator.Service,
) PostEventsCustoms {
	return func(c *fiber.Ctx) error {
		if utils.UnsafeString(c.Request().Header.ContentType()) != fiber.MIMEApplicationJSON {
			return fiber.NewError(fiber.StatusBadRequest, "content type is not application/json")
		}

		customEv := event.Custom{}
		customEv.Timestamp = time.Now().UTC()
		customEv.Name = utils.CopyString(c.Params("name"))

		// Referrer of the POST request, that is the viewed page.
		requestReferrer := peekReferrerHeader(c)
		err := customEv.PageUri.Parse(requestReferrer)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Referer or X-Prisme-Referrer")
		}

		// Validate properties.
		body := utils.CopyBytes(c.Body())
		if len(body) > 0 {
			result := gjson.GetManyBytes(utils.CopyBytes(c.Body()), "@keys", "@values")
			result[0].ForEach(func(_, keys gjson.Result) bool {
				customEv.Keys = append(customEv.Keys, keys.String())
				return true
			})
			result[1].ForEach(func(_, values gjson.Result) bool {
				customEv.Values = append(customEv.Values, values.String())
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
