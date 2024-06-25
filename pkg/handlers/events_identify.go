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
	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
)

type PostEventsIdentify fiber.Handler

// ProvidePostEventsIdentify is a wire provider for POST /api/v1/events/identify.
// handler.
func ProvidePostEventsIdentify(
	logger zerolog.Logger,
	eventStore eventstore.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
) PostEventsIdentify {
	return func(c *fiber.Ctx) error {
		// ContentType must be json.
		if utils.UnsafeString(c.Request().Header.ContentType()) != fiber.MIMEApplicationJSON {
			return fiber.NewError(fiber.StatusBadRequest, "content type is not application/json")
		}

		// Referrer of the POST request, that is the viewed page.
		requestReferrer := peekReferrerHeader(c)

		identifyEvent := event.Identify{}

		// Parse page URI.
		pageUri, err := uri.ParseBytes(requestReferrer)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Referer or X-Prisme-Referrer")
		}

		// Compute device id.
		deviceId := computeDeviceId(
			saltManagerService.StaticSalt().Bytes(), c.Request().Header.UserAgent(),
			utils.UnsafeBytes(c.IP()), utils.UnsafeBytes(pageUri.Host()),
		)

		// Check if visitor id must be updated.
		result := gjson.GetBytes(c.Body(), "visitorId")
		if result.Exists() {
			if result.Str == "" {
				return fiber.NewError(fiber.StatusBadRequest, "visitorId must be a string or undefined")
			}

			var ok bool
			identifyEvent.Session, ok = sessionStorage.IdentifySession(deviceId, result.Str)
			// Session not found.
			if !ok {
				return errSessionNotFound
			}
		} else {
			var ok bool
			identifyEvent.Session, ok = sessionStorage.GetSession(deviceId)
			if !ok {
				return errSessionNotFound
			}
		}

		// Retrieve set once properties.
		result = gjson.GetBytes(c.Body(), "setOnce")
		if result.Exists() {
			collectJsonKeyValues(
				utils.UnsafeBytes(result.Raw),
				&identifyEvent.InitialKeys,
				&identifyEvent.InitialValues,
			)
		}

		// Retrive other properties.
		result = gjson.GetBytes(c.Body(), "set")
		if result.Exists() {
			collectJsonKeyValues(
				utils.UnsafeBytes(result.Raw),
				&identifyEvent.Keys,
				&identifyEvent.Values,
			)
		}

		identifyEvent.Timestamp = time.Now().UTC()
		err = eventStore.StoreIdentifyEvent(c.UserContext(), &identifyEvent)
		if err != nil {
			return fmt.Errorf("failed to store identify event: %w", err)
		}

		return nil
	}
}
