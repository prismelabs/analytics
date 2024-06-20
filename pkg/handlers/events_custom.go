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

		// Parse page URI.
		err := customEv.PageUri.Parse(requestReferrer)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Referer or X-Prisme-Referrer")
		}

		// Compute device id.
		deviceId := computeDeviceId(
			saltManagerService.StaticSalt().Bytes(), c.Request().Header.UserAgent(),
			utils.UnsafeBytes(c.IP()), customEv.PageUri.Host(),
		)

		var ok bool
		customEv.Session, ok = sessionStorage.GetSession(deviceId)
		// Session not found.
		if !ok {
			return fiber.NewError(fiber.StatusBadRequest, "session not found")
		}

		// Event date and name.
		customEv.Timestamp = time.Now().UTC()
		customEv.Name = utils.CopyString(c.Params("name"))

		// Validate properties.
		if len(c.Body()) > 0 {
			collectJsonKeyValues(c.Body(), &customEv.Keys, &customEv.Values)
		}

		// Store event.
		err = eventStore.StoreCustom(c.UserContext(), &customEv)
		if err != nil {
			return fmt.Errorf("failed to store custom event: %w", err)
		}

		return nil
	}
}
