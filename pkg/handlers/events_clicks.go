package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/dataview"
	"github.com/prismelabs/analytics/pkg/event"
	hutils "github.com/prismelabs/analytics/pkg/handlers/utils"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstorage"
)

type PostEventsClicks fiber.Handler

// ProvidePostEventsClicks is a wire provider for POST /api/v1/events/clicks
// handler.
func ProvidePostEventsClicks(
	eventStore eventstore.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
) PostEventsClicks {
	return func(c *fiber.Ctx) error {
		var err error
		clickEv := event.Click{}

		// ContentType must be json if request has a body.
		if utils.UnsafeString(c.Request().Header.ContentType()) != fiber.MIMEApplicationJSON {
			return fiber.NewError(fiber.StatusBadRequest, "content type is not application/json")
		}

		// Parse referrer.
		clickEv.PageUri, err = hutils.PeekAndParseReferrerHeader(c)
		if err != nil {
			return err
		}

		// Create body JSON view.
		jsonData := dataview.NewJsonData(c.Body())
		jsonView := dataview.JsonKvView{Json: jsonData}

		// Extract event tag and id.
		tag, err := jsonView.GetString("tag")
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		id, err := jsonView.GetString("id")
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		// Compute device id.
		deviceId := hutils.ComputeDeviceId(
			saltManagerService.StaticSalt().Bytes(), c.Request().Header.UserAgent(),
			utils.UnsafeBytes(c.IP()), utils.UnsafeBytes(clickEv.PageUri.Host()),
		)

		// Retrieve visitor session.
		ctx := c.UserContext()
		var ok bool
		clickEv.Session, ok = sessionStorage.WaitSession(deviceId, clickEv.PageUri, hutils.ContextTimeout(ctx))
		if !ok {
			return errSessionNotFound
		}

		// Add event data.
		clickEv.Timestamp = time.Now().UTC()
		clickEv.Tag = tag
		clickEv.Id = id

		// Store event.
		err = eventStore.StoreClick(ctx, &clickEv)
		if err != nil {
			return fmt.Errorf("failed to store custom event: %w", err)
		}

		return nil
	}
}
