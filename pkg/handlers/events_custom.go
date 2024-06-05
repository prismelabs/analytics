package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstorage"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
)

type PostEventsCustom fiber.Handler

// ProvidePostEventsCustom is a wire provider for POST /api/v1/events/custom/:name events handler.
func ProvidePostEventsCustom(
	logger zerolog.Logger,
	eventStore eventstore.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
) PostEventsCustom {
	return func(c *fiber.Ctx) error {
		// ContentType must be json if request has a body.
		if c.Request().Header.ContentLength() != 0 &&
			utils.UnsafeString(c.Request().Header.ContentType()) != fiber.MIMEApplicationJSON {
			return fiber.NewError(fiber.StatusBadRequest, "content type is not application/json")
		}

		// Referrer of the POST request, that is the viewed page.
		requestReferrer := peekReferrerHeader(c)

		customEv := event.Custom{}

		// Event date and name.
		customEv.Timestamp = time.Now().UTC()
		customEv.Name = utils.CopyString(c.Params("name"))

		// Parse page URI.
		err := customEv.PageUri.Parse(requestReferrer)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Referer or X-Prisme-Referrer")
		}

		userAgent := c.Request().Header.UserAgent()
		customEv.Session.VisitorId = computeVisitorId(
			userAgent, saltManagerService.DailySalt().Bytes(),
			utils.UnsafeBytes(c.IP()), customEv.PageUri.Host(),
		)

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

		var ok bool
		customEv.Session, ok = sessionStorage.GetSession(customEv.Session.VisitorId)
		// Session not found.
		// This can happen if tracking script is not installed on all pages or
		// prisme instance was restarted.
		if !ok {
			return fiber.NewError(fiber.StatusBadRequest, "session not found")
		}

		// Store event.
		err = eventStore.StoreCustom(c.UserContext(), &customEv)
		if err != nil {
			return fmt.Errorf("failed to store custom event: %w", err)
		}

		return nil
	}
}
