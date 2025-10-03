//go:build test

package faker

import (
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/prismelabs/analytics/pkg/uri"
)

const (
	AlphaLower = "abcdefghijklmnopqrstuvwxyz"
	AlphaUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Alpha      = AlphaLower + AlphaUpper
	Num        = "0123456789"
	AlphaNum   = Alpha + Num
)

// Item returns a random item within provided slice. This function panics if the
// slice is empty.
func Item[T any](slice []T) T {
	index := rand.Int() % len(slice)
	return slice[index]
}

// String generates a random string of length `length` using provided `charset`.
func String(charset string, length int) string {
	buf := make([]byte, length)

	for i := range length {
		buf[i] = charset[rand.Intn(len(charset)-1)]
	}

	return utils.UnsafeString(buf)
}

// CountryCode generates an ISO 3166-1 alpha-2 code.
func CountryCode() ipgeolocator.CountryCode {
	return ipgeolocator.NewCountryCode(Item(CountryCodesList))
}

// UserAgent returns a random and popular user agent (desktop or mobile).
func UserAgent() string {
	if rand.Int()%2 == 0 {
		return DesktopUserAgent()
	}

	return MobileUserAgent()
}

// DesktopUserAgent returns a random and popular desktop user agent.
func DesktopUserAgent() string {
	return Item(DesktopUserAgents)
}

// MobileUserAgent returns a random and popular mobile user agent.
func MobileUserAgent() string {
	return Item(MobileUserAgents)
}

// UapDesktopClient returns a random uaparser.Client.
func UapClient() uaparser.Client {
	if rand.Int()%2 == 0 {
		return UapDesktopClient()
	}

	return UapMobileClient()
}

// UapDesktopClient returns a random desktop uaparser.Client.
func UapDesktopClient() uaparser.Client {
	return Item(DesktopClients)
}

// UapMobileClient returns a random mobile uaparser.Client.
func UapMobileClient() uaparser.Client {
	return Item(MobileClients)
}

func Domain() string {
	return Item(DomainList)
}

// Path returns a random path.
func Path() string {
	return Item(PathList)
}

// Uri returns a random HTTP URI with a scheme, a domain, and path.
func Uri() uri.Uri {
	scheme := Item([]string{"http://", "https://"})
	domain := Domain()
	path := Path()
	return testutils.Must(uri.Parse)(scheme + domain + path)
}

// PageUri returns a random page URI associated with given session.
func PageUri(session event.Session) uri.Uri {
	return testutils.Must(uri.Parse)(
		session.PageUri.Scheme() + "://" + session.PageUri.Host() + Item(PathList),
	)
}

// Time returns a random time between now + d. If d is negative return time
// maybe in the past.
func Time(d time.Duration) time.Time {
	now := time.Now()
	if int(d) < 0 {
		return now.Add(
			-time.Duration(
				rand.Intn(int(-d)),
			),
		)
	}

	return now.Add(time.Duration(rand.Intn(int(d))))
}
