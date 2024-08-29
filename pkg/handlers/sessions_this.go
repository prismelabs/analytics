package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstorage"
	"github.com/prismelabs/analytics/pkg/uri"
)

type GetSessionsThis fiber.Handler

// ProvideGetSessionsThis is a wire provider for GET /api/v1/sessions/@this.
func ProvideGetSessionsThis(
	saltManagerService saltmanager.Service,
	sessions sessionstorage.Service,
) GetSessionsThis {
	return func(c *fiber.Ctx) error {
		// Allow all origin has this handler is executed after origin registry
		// middleware.
		c.Response().Header.Add("Access-Control-Allow-Origin", "*")

		userAgent := c.Context().UserAgent()
		ipAddr := utils.UnsafeBytes(c.IP())
		requestReferrer := peekReferrerHeader(c)
		pageUri, err := uri.ParseBytes(requestReferrer)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Referer or X-Prisme-Referrer")
		}

		// Compute device id.
		deviceId := computeDeviceId(
			saltManagerService.StaticSalt().Bytes(), userAgent,
			ipAddr, utils.UnsafeBytes(pageUri.Host()),
		)

		session, ok := sessions.WaitSession(deviceId, 0)
		if !ok {
			return errSessionNotFound
		}

		return c.JSON(session)
	}
}
