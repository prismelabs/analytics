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

type PostEventsClicksFileDownload fiber.Handler

// ProvidePostEventsClicksFileDownload is a wire provider for POST
// /api/v1/events/clicks/file-download handler.
func ProvidePostEventsClicksFileDownload(
	eventStore eventstore.Service,
	saltManagerService saltmanager.Service,
	sessionStorage sessionstorage.Service,
) PostEventsClicksFileDownload {
	return func(c *fiber.Ctx) error {
		var err error
		FileDownloadEv := event.FileDownload{}

		fileUri, err := uri.ParseBytes(c.Body())
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid outbound link: %v", err.Error()))
		}

		// Parse referrer.
		FileDownloadEv.PageUri, err = hutils.PeekAndParseReferrerHeader(c)
		if err != nil {
			return err
		}

		// Compute device id.
		deviceId := hutils.ComputeDeviceId(
			saltManagerService.StaticSalt().Bytes(), c.Request().Header.UserAgent(),
			utils.UnsafeBytes(c.IP()), utils.UnsafeBytes(FileDownloadEv.PageUri.Host()),
		)

		// Retrieve visitor session.
		ctx := c.UserContext()
		var ok bool
		FileDownloadEv.Session, ok = sessionStorage.WaitSession(deviceId, FileDownloadEv.PageUri, hutils.ContextTimeout(ctx))
		if !ok {
			return errSessionNotFound
		}

		// Add event data.
		FileDownloadEv.Timestamp = time.Now().UTC()
		FileDownloadEv.FileUrl = fileUri

		// Store event.
		err = eventStore.StoreFileDownload(ctx, &FileDownloadEv)
		if err != nil {
			return fmt.Errorf("failed to store custom event: %w", err)
		}

		return nil
	}
}
