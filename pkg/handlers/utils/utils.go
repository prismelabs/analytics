package utils

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

// BodyOrEmptyJsonObj returns request body or an empty JSON object buffer
// otherwise.
func BodyOrEmptyJsonObj(c *fiber.Ctx) []byte {
	body := c.Body()
	if len(body) == 0 {
		body = emptyJsonObj
	}

	return body
}

// PeekReferrerHeader peek X-Prisme-Referrer header and fallback to
// standard Referer header otherwise.
func PeekReferrerHeader(c *fiber.Ctx) []byte {
	referrer := c.Request().Header.Peek("X-Prisme-Referrer")

	// No X-Prisme-Referrer header, javascript is probably disabled.
	if len(referrer) == 0 {
		// Fallback to standard referer header (with its limitation).
		referrer = c.Request().Header.Peek(fiber.HeaderReferer)
	}

	return referrer
}

// PeekReferrerQueryOrHeader peek referrer from "referrer" query parameter
// and fallback to standard Referer header otherwise.
func PeekReferrerQueryOrHeader(c *fiber.Ctx) []byte {
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

// ComputeDeviceId computes xxh3 hash of the given byte slices.
// This is the same as Xxh3 function.
func ComputeDeviceId(bytesSlice ...[]byte) uint64 {
	return Xxh3(bytesSlice...)
}

// ComputeVisitorId computes xxh3 hash of the given byte slices and
// adds result as hexadecimal suffix to "prisme_".
func ComputeVisitorId(bytesSlice ...[]byte) string {
	return fmt.Sprintf("prisme_%X", Xxh3(bytesSlice...))
}

// Xxh3 computes xxh3 hash of the given byte slices.
func Xxh3(bytesSlice ...[]byte) uint64 {
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

// ExtractUtmParams extracts UTM parameters from the given query args.
// If no utm_source arg is found, it fallbacks to ref arg.
func ExtractUtmParams(args *fasthttp.Args) event.UtmParams {
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

// ContextTimeout extract duration until context timeout and panics
// if no deadline is found.
func ContextTimeout(ctx context.Context) time.Duration {
	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		panic("context has no deadline")
	}

	return time.Until(deadline)
}
