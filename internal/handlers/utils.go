package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// mustRender render template and sets the given status code. If an error occured
// this function panic.
func mustRender(c *fiber.Ctx, statusCode int, name string, bind interface{}, layouts ...string) {
	c.Response().SetStatusCode(statusCode)
	err := c.Render(name, bind, layouts...)
	if err != nil {
		panic(err)
	}
}

func peekReferrerHeader(c *fiber.Ctx) []byte {
	referrer := c.Request().Header.Peek("X-Prisme-Referrer")

	// No X-Prisme-Referrer header, javascript is probably disabled.
	if len(referrer) == 0 {
		// Fallback to standard referer header (with its limitation).
		referrer = c.Request().Header.Peek(fiber.HeaderReferer)
	}

	return referrer
}
