package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/uri"
)

// ReferrerAsDefaultOrigin returns a middleware that sets request origin to
// referrer header if undefined.
func ReferrerAsDefaultOrigin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		headers := &c.Request().Header

		// Origin is missing. This can happen on cross origin GET requests.
		if len(headers.Peek(fiber.HeaderOrigin)) == 0 {
			referrer, err := uri.ParseBytes(headers.Peek(fiber.HeaderReferer))
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, `invalid "Referer"`)
			}

			headers.Set(fiber.HeaderOrigin, referrer.Origin())
		}

		return c.Next()
	}
}
