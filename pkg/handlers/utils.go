package handlers

import (
	"fmt"

	"github.com/cespare/xxhash/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
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

func equalBytes(a, b []byte) bool {
	return utils.UnsafeString(a) == utils.UnsafeString(b)
}

func computeDeviceId(bytesSlice ...[]byte) string {
	return fmt.Sprintf("%X", xxh3(bytesSlice...))
}

func computeVisitorId(prefix string, bytesSlice ...[]byte) string {
	return fmt.Sprintf("%v%X", prefix, xxh3(bytesSlice...))
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

func collectJsonKeyValues(json []byte, keys, values *[]string) {
	// Get keys.
	result := gjson.GetBytes(json, "@keys")
	result.ForEach(func(_, key gjson.Result) bool {
		*keys = append(*keys, utils.CopyString(key.String()))
		return true
	})

	// Get values.
	result = gjson.GetBytes(json, "@values")
	result.ForEach(func(_, value gjson.Result) bool {
		*values = append(*values, utils.CopyString(value.Raw))
		return true
	})
}
