package handlers

import (
	"encoding/binary"
	"fmt"

	"github.com/cespare/xxhash/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/event"
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

func equalBytes(a, b []byte) bool {
	return utils.UnsafeString(a) == utils.UnsafeString(b)
}

func computeVisitorId(bytesSlice ...[]byte) string {
	return fmt.Sprintf("prisme_%X", xxh3(bytesSlice...))
}

func xxh3(bytesSlice ...[]byte) uint64 {
	hash := xxhash.New()

	for _, slice := range bytesSlice {
		_, err := hash.Write(slice)
		// Should never happen as documented in hash.Write.
		if err != nil {
			panic(err)
		}
	}

	return hash.Sum64()
}

func sessionKey(visitorId string) string {
	return fmt.Sprintf("session_id[%q]", visitorId)
}

func computeSessionId(pageView *event.PageView) uint64 {
	return xxh3(
		binary.LittleEndian.AppendUint64(nil, uint64(pageView.Timestamp.UnixNano())),
		utils.UnsafeBytes(pageView.VisitorId),
		pageView.PageUri.Host(),
	)
}
