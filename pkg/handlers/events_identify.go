package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/dataview"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstorage"
	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/rs/zerolog"
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

		return eventIdentifyHandler(
			c.UserContext(),
			logger,
			eventStore,
			saltManagerService,
			sessionStorage,
			requestReferrer,
			c.Request().Header.UserAgent(),
			utils.UnsafeBytes(c.IP()),
			dataview.JsonKvView{Json: c.Body()},
			dataview.JsonKvCollector{Json: c.Body(), Path: "set."},
			dataview.JsonKvCollector{Json: c.Body(), Path: "setOnce."},
		)
	}
}

func eventIdentifyHandler(
	ctx context.Context,
	_ zerolog.Logger,
	eventStore eventstore.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
	requestReferrer, userAgent, ipAddr []byte,
	kvView dataview.KvView, setPropCollector, setOncePropCollector dataview.KvCollector,
) error {

	identifyEvent := event.Identify{}

	// Parse page URI.
	pageUri, err := uri.ParseBytes(requestReferrer)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid Referer or X-Prisme-Referrer")
	}

	// Compute device id.
	deviceId := computeDeviceId(
		saltManagerService.StaticSalt().Bytes(), userAgent,
		ipAddr, utils.UnsafeBytes(pageUri.Host()),
	)

	// Retrieve visitor ID.
	visitorId, err := kvView.GetString("visitorId")
	if err != nil && !errors.Is(err, dataview.ErrKvViewEntryNotFound) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// No visitor ID provided.
	if err != nil && errors.Is(err, dataview.ErrKvViewEntryNotFound) {
		var ok bool
		identifyEvent.Session, ok = sessionStorage.WaitSession(deviceId, contextTimeout(ctx))
		if !ok {
			return errSessionNotFound
		}
	} else { // Visitor id provided.
		var ok bool
		identifyEvent.Session, ok = sessionStorage.IdentifySession(deviceId, visitorId)
		// Session not found.
		if !ok {
			// Wait for it.
			_, ok = sessionStorage.WaitSession(deviceId, contextTimeout(ctx))
			if !ok {
				return errSessionNotFound
			}
			identifyEvent.Session, ok = sessionStorage.IdentifySession(deviceId, visitorId)
			if !ok {
				return errSessionNotFound
			}
		}
	}

	// Collect properties.
	identifyEvent.InitialKeys, identifyEvent.InitialValues = setOncePropCollector.CollectKeysValues()
	identifyEvent.Keys, identifyEvent.Values = setPropCollector.CollectKeysValues()

	identifyEvent.Timestamp = time.Now().UTC()
	err = eventStore.StoreIdentifyEvent(ctx, &identifyEvent)
	if err != nil {
		return fmt.Errorf("failed to store identify event: %w", err)
	}

	return nil
}
