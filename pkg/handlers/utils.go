package handlers

import (
	"fmt"

	"github.com/cespare/xxhash/v2"
	"github.com/gofiber/fiber/v2"
)

func peekReferrerHeader(c *fiber.Ctx) []byte {
	referrer := c.Request().Header.Peek("X-Prisme-Referrer")

	// No X-Prisme-Referrer header, javascript is probably disabled.
	if len(referrer) == 0 {
		// Fallback to standard referer header (with its limitation).
		referrer = c.Request().Header.Peek(fiber.HeaderReferer)
	}

	return referrer
}

func computeVisitorId(bytesSlice ...[]byte) string {
	hash := xxhash.New()

	for _, slice := range bytesSlice {
		_, err := hash.Write(slice)
		// Should never happen as documented in hash.Write.
		if err != nil {
			panic(err)
		}
	}

	return fmt.Sprintf("prisme_%X", hash.Sum64())
}
