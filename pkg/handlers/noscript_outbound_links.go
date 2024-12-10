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

type GetNoscriptEventsOutboundLinks fiber.Handler

// ProvideGetNoscriptEventsOutboundLinks is a wire provider for
// GET /api/v1/noscript/events/outbound-links handler.
func ProvideGetNoscriptEventsOutboundLinks(
	eventStore eventstore.Service,
	sessionStorage sessionstorage.Service,
	saltManagerService saltmanager.Service,
) GetNoscriptEventsOutboundLinks {
	return func(c *fiber.Ctx) error {
		var err error
		outboundLinkClickEv := event.OutboundLinkClick{}

		outboundLinkClickEv.PageUri, err = hutils.PeekAndParseReferrerQueryHeader(c)
		if err != nil {
			return err
		}

		outboundUri, err := uri.Parse(c.Query("url"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid outbound link: %v", err.Error()))
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

		ctx := c.UserContext()
		var ok bool
		outboundLinkClickEv.Session, ok = sessionStorage.WaitSession(deviceId, outboundLinkClickEv.PageUri, hutils.ContextTimeout(ctx))
		if !ok {
			// Fallback to root of referrer. This is needed if referrer query or header contained entire url
			// while referrer pageview event contains only origin because of different referrer policy.
			outboundLinkClickEv.PageUri = outboundLinkClickEv.PageUri.RootUri()
			outboundLinkClickEv.Session, ok = sessionStorage.WaitSession(deviceId, outboundLinkClickEv.PageUri, hutils.ContextTimeout(ctx))
		}
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

		return c.Redirect(outboundUri.String(), fiber.StatusFound)
	}
}
