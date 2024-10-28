package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/event"
	hutils "github.com/prismelabs/analytics/pkg/handlers/utils"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstorage"
	"github.com/prismelabs/analytics/pkg/uri"
)

type PostEventsClicksOutboundLink fiber.Handler

// ProvidePostEventsClicksOutboundLink is a wire provider for POST
// /api/v1/events/clicks/outbound-link handler.
func ProvidePostEventsClicksOutboundLink(
	eventStore eventstore.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
) PostEventsClicksOutboundLink {
	return func(c *fiber.Ctx) error {
		var err error
		outboundLinkClickEv := event.OutboundLinkClick{}

		outboundUri, err := uri.ParseBytes(c.Body())
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid outbound link: %v", err.Error()))
		}

		// Parse referrer.
		outboundLinkClickEv.PageUri, err = hutils.PeekAndParseReferrerHeader(c)
		if err != nil {
			return err
		}

		// Check that link is external.
		if outboundUri.Host() == outboundLinkClickEv.PageUri.Host() {
			return fiber.NewError(fiber.StatusBadRequest, "internal link")
		}

		// Compute device id.
		deviceId := hutils.ComputeDeviceId(
			saltManagerService.StaticSalt().Bytes(), c.Request().Header.UserAgent(),
			utils.UnsafeBytes(c.IP()), utils.UnsafeBytes(outboundLinkClickEv.PageUri.Host()),
		)

		// Retrieve visitor session.
		ctx := c.UserContext()
		var ok bool
		outboundLinkClickEv.Session, ok = sessionStorage.WaitSession(deviceId, outboundLinkClickEv.PageUri, hutils.ContextTimeout(ctx))
		if !ok {
			return errSessionNotFound
		}

		// Add event data.
		outboundLinkClickEv.Timestamp = time.Now().UTC()
		outboundLinkClickEv.Link = outboundUri

		// Store event.
		err = eventStore.StoreOutboundLinkClick(ctx, &outboundLinkClickEv)
		if err != nil {
			return fmt.Errorf("failed to store custom event: %w", err)
		}

		return nil
	}
}