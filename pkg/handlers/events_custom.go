package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/dataview"
	"github.com/prismelabs/analytics/pkg/event"
	hutils "github.com/prismelabs/analytics/pkg/handlers/utils"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstore"
	"github.com/prismelabs/analytics/pkg/uri"
)

type PostEventsCustom fiber.Handler

// ProvidePostEventsCustom is a wire provider for POST /api/v1/events/custom/:name handler.
func ProvidePostEventsCustom(
	eventStore eventstore.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstore.Service,
) PostEventsCustom {
	return func(c *fiber.Ctx) error {

		// ContentType must be json if request has a body.
		if c.Request().Header.ContentLength() != 0 &&
			utils.UnsafeString(c.Request().Header.ContentType()) != fiber.MIMEApplicationJSON {
			return fiber.NewError(fiber.StatusBadRequest, "content type is not application/json")
		}

		// Parse referrer.
		referrer, err := hutils.PeekAndParseReferrerHeader(c)
		if err != nil {
			return err
		}

		data := dataview.NewJsonData(hutils.BodyOrEmptyJsonObj(c))
		return eventsCustomHandler(
			c.UserContext(),
			eventStore,
			saltManagerService,
			sessionStorage,
			referrer,
			c.Request().Header.UserAgent(),
			utils.UnsafeBytes(c.IP()),
			c.Params("name"),
			dataview.JsonKvCollector{Json: data},
		)
	}
}

func eventsCustomHandler(
	ctx context.Context,
	eventStore eventstore.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstore.Service,
	requestReferrer uri.Uri,
	userAgent, ipAddr []byte,
	eventName string,
	kvCollector dataview.KvCollector,
) (err error) {
	customEv := event.Custom{
		PageUri: requestReferrer,
	}

	// Compute device id.
	deviceId := hutils.ComputeDeviceId(
		saltManagerService.StaticSalt().Bytes(), userAgent,
		ipAddr, utils.UnsafeBytes(customEv.PageUri.Host()),
	)

	var ok bool
	customEv.Session, ok = sessionStorage.WaitSession(deviceId, customEv.PageUri, hutils.ContextTimeout(ctx))
	// Session not found.
	if !ok {
		return errSessionNotFound
	}

	// Event date and name.
	customEv.Timestamp = time.Now().UTC()
	customEv.Name = utils.CopyString(eventName)

	// Collect event properties.
	customEv.Keys, customEv.Values, err = kvCollector.CollectKeysValues()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Store event.
	err = eventStore.StoreCustom(ctx, &customEv)
	if err != nil {
		return fmt.Errorf("failed to store custom event: %w", err)
	}

	return nil
}
