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

type PostEventsFileDownloads fiber.Handler

// ProvidePostEventsFileDownloads is a wire provider for POST
// /api/v1/events/file-downloads handler.
func ProvidePostEventsFileDownloads(
	eventStore eventstore.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
) PostEventsFileDownloads {
	return func(c *fiber.Ctx) error {
		var err error
		fileDownloadEv := event.FileDownload{}

		var fileUri uri.Uri
		isPing := utils.UnsafeString(c.Body()) == "PING"

		// Ping attribute of HTML anchor element.
		if isPing {
			// Parse URI of visitor pages.
			fileDownloadEv.PageUri, err = uri.ParseBytes(c.Request().Header.Peek(fiber.HeaderPingFrom))
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid Ping-From header: %v", err.Error()))
			}

			// Parse URI of downloaded file.
			fileUri, err = uri.ParseBytes(c.Request().Header.Peek(fiber.HeaderPingTo))
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid Ping-To header: %v", err.Error()))
			}
		} else {
			// Parse referrer.
			fileDownloadEv.PageUri, err = hutils.PeekAndParseReferrerHeader(c)
			if err != nil {
				return err
			}

			// Parse URI of downloaded file.
			fileUri, err = uri.ParseBytes(c.Body())
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid outbound link: %v", err.Error()))
			}
		}

		// Compute device id.
		deviceId := hutils.ComputeDeviceId(
			saltManagerService.StaticSalt().Bytes(), c.Request().Header.UserAgent(),
			utils.UnsafeBytes(c.IP()), utils.UnsafeBytes(fileDownloadEv.PageUri.Host()),
		)

		// Retrieve visitor session.
		ctx := c.UserContext()
		var ok bool
		fileDownloadEv.Session, ok = sessionStorage.WaitSession(deviceId, fileDownloadEv.PageUri, hutils.ContextTimeout(ctx))
		if !ok && isPing {
			// Fallback to root of referrer. This is needed as Ping-From contains entire url
			// while referrer header may only contains origin depending on referrer policy.
			fileDownloadEv.PageUri = fileDownloadEv.PageUri.RootUri()
			fileDownloadEv.Session, ok = sessionStorage.WaitSession(deviceId, fileDownloadEv.PageUri, hutils.ContextTimeout(ctx))
		}
		if !ok {
			return errSessionNotFound
		}

		// Add event data.
		fileDownloadEv.Timestamp = time.Now().UTC()
		fileDownloadEv.FileUrl = fileUri

		// Store event.
		err = eventStore.StoreFileDownload(ctx, &fileDownloadEv)
		if err != nil {
			return fmt.Errorf("failed to store custom event: %w", err)
		}

		return nil
	}
}
