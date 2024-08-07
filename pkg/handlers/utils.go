package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/valyala/fasthttp"
)

var (
	emptyJsonObj = []byte{'{', '}'}
)

func bodyOrEmptyJsonObj(c *fiber.Ctx) []byte {
	body := c.Body()
	if len(body) == 0 {
		body = emptyJsonObj
	}

	return body
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

func peekReferrerQueryOrHeader(c *fiber.Ctx) []byte {
	referrer := utils.UnsafeBytes(c.Query("referrer"))
	if len(referrer) == 0 {
		referrer = c.Request().Header.Peek("X-Prisme-Referrer")
	}

	// No X-Prisme-Referrer header, javascript is probably disabled.
	if len(referrer) == 0 {
		// Fallback to standard referer header (with its limitation).
		referrer = c.Request().Header.Peek(fiber.HeaderReferer)
	}

	return referrer
}

func computeDeviceId(bytesSlice ...[]byte) uint64 {
	return xxh3(bytesSlice...)
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

func extractUtmParams(args *fasthttp.Args) event.UtmParams {
	utmParams := event.UtmParams{}
	if args.Len() == 0 {
		return utmParams
	}

	utmParams.Source = string(args.Peek("utm_source"))
	if utmParams.Source == "" {
		utmParams.Source = string(args.Peek("ref"))
	}

	utmParams.Medium = string(args.Peek("utm_medium"))
	utmParams.Campaign = string(args.Peek("utm_campaign"))
	utmParams.Term = string(args.Peek("utm_term"))
	utmParams.Content = string(args.Peek("utm_content"))

	return utmParams
}

func contextTimeout(ctx context.Context) time.Duration {
	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		panic("context has no deadline")
	}

	return time.Until(deadline)
}
