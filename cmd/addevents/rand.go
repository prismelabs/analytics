package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
)

func randomDesktopClient() uaparser.Client {
	return randomItem(desktopClients)
}

func randomMobileClient() uaparser.Client {
	return randomItem(mobileClients)
}

func randomPathName() string {
	return randomItem(pathnamesList)
}

func randomExternalDomain() string {
	return randomItem(externalDomainsList)
}

func randomMinute() time.Duration {
	return time.Duration((rand.Int() % 60)) * time.Minute
}

func randomItem[T any](slice []T) T {
	index := rand.Int() % len(slice)
	return slice[index]
}

const (
	alphaLower = "abcdefghijklmnopqrstuvwxyz"
	alphaUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alpha      = alphaLower + alphaUpper
	num        = "0123456789"
	alphaNum   = alpha + num
)

func randomString(charset string, length int) string {
	buf := make([]byte, length)

	for i := range length {
		buf[i] = charset[rand.Intn(len(charset)-1)]
	}

	return utils.UnsafeString(buf)
}

func randomVisitorId(idsRange uint64) string {
	return fmt.Sprintf("prisme_%X", rand.Uint64()%idsRange)
}

func randomCountryCode() string {
	return randomItem(countryCodesList)
}

func randomCustomEvent(session Session) CustomEvent {
	customEv := CustomEvent{
		session: session,
	}

	customEv.name = randomItem([]string{"click", "download", "sign_up", "subscription", "lot_of_props"})
	switch customEv.name {
	case "click":
		customEv.keys = []string{"x", "y"}
		customEv.values = []string{fmt.Sprint(rand.Intn(3000)), fmt.Sprint(rand.Intn(2000))}

	case "download":
		customEv.keys = []string{"doc"}
		customEv.values = []string{fmt.Sprintf("%v.pdf", randomString(alphaLower, 3))}

	case "sign_up":
		customEv.keys = []string{}
		customEv.values = []string{}

	case "subscription":
		customEv.keys = []string{"plan"}
		customEv.values = []string{randomItem([]string{"growth", "premium", "enterprise"})}

	case "lot_of_props":

		customEv.keys = make([]string, 64)
		customEv.values = make([]string, len(customEv.keys))
		for i := 0; i < len(customEv.keys); i++ {
			customEv.keys[i] = randomString(alphaLower, 3)
			customEv.values[i] = randomString(alpha, 9)
		}

	default:
		panic("not implemented")
	}

	return customEv
}

var (
	pathnamesList = []string{
		"/foo/bar/qux",
		"/foo/bar/",
		"/foo/bar",
		"/foo",
		"/blog",
		"/blog/misc/a-nice-post",
		"/blog/misc/another-nice-post",
		"/contact",
		"/terms-of-service",
		"/privacy",
	}
	externalDomainsList = []string{
		"youtube.com", "www.google.com", "www.blogger.com", "cloudflare.com",
		"apple.com", "linkedin.com", "googleusercontent.com", "support.google.com",
		"docs.google.com", "microsoft.com", "wordpress.org", "whatsapp.com",
		"maps.google.com", "play.google.com", "en.wikipedia.org", "plus.google.com",
		"sites.google.com", "accounts.google.com", "t.me", "europa.eu",
		"bp.blogspot.com", "youtu.be", "drive.google.com", "mozilla.org", "vk.com",
		"adobe.com", "facebook.com", "vimeo.com", "uol.com.br",
		"policies.google.com", "g.page", "istockphoto.com", "amazon.com",
		"github.com", "paypal.com", "tiktok.com", "tools.google.com",
		"brandbucket.com", "ok.ru", "gravatar.com", "wikimedia.org",
		"news.google.com", "opera.com", "dropbox.com", "feedburner.com",
		"netvibes.com", "files.wordpress.com", "draft.blogger.com",
		"googleblog.com", "myspace.com",
	}
	countryCodesList = []string{
		"AF", "AX", "AL", "DZ", "AS", "AD", "AO", "AI", "AQ", "AG", "AR", "AM", "AW",
		"AU", "AT", "AZ", "BS", "BH", "BD", "BB", "BY", "BE", "BZ", "BJ", "BM", "BT",
		"BO", "BQ", "BA", "BW", "BV", "BR", "IO", "BN", "BG", "BF", "BI", "CV",
		"KH", "CM", "CA", "KY", "CF", "TD", "CL", "CN", "CX", "CC", "CO", "KM", "CD",
		"CG", "CK", "CR", "CI", "HR", "CU", "CW", "CY", "CZ", "DK", "DJ", "DM", "DO",
		"EC", "EG", "SV", "GQ", "ER", "EE", "SZ", "ET", "FK", "FO", "FJ", "FI", "FR",
		"GF", "PF", "TF", "GA", "GM", "GE", "DE", "GH", "GI", "GR", "GL", "GD", "GP",
		"GU", "GT", "GG", "GN", "GW", "GY", "HT", "HM", "VA", "HN", "HK", "HU", "IS",
		"IN", "ID", "IR", "IQ", "IE", "IM", "IL", "IT", "JM", "JP", "JE", "JO", "KZ",
		"KE", "KI", "KP", "KR", "KW", "KG", "LA", "LV", "LB", "LS", "LR", "LY", "LI",
		"LT", "LU", "MO", "MG", "MW", "MY", "MV", "ML", "MT", "MH", "MQ", "MR", "MU",
		"YT", "MX", "FM", "MD", "MC", "MN", "ME", "MS", "MA", "MZ", "MM", "NA", "NR",
		"NP", "NL", "NC", "NZ", "NI", "NE", "NG", "NU", "NF", "MK", "MP", "NO", "OM",
		"PK", "PW", "PS", "PA", "PG", "PY", "PE", "PH", "PN", "PL", "PT", "PR", "QA",
		"RE", "RO", "RU", "RW", "BL", "SH", "KN", "LC", "MF", "PM", "VC", "WS", "SM",
		"ST", "SA", "SN", "RS", "SC", "SL", "SG", "SX", "SK", "SI", "SB", "SO", "ZA",
		"GS", "SS", "ES", "LK", "SD", "SR", "SJ", "SE", "CH", "SY", "TW", "TJ", "TZ",
		"TH", "TL", "TG", "TK", "TO", "TT", "TN", "TR", "TM", "TC", "TV", "UG", "UA",
		"AE", "GB", "UM", "US", "UY", "UZ", "VU", "VE", "VN", "VG", "VI", "WF", "EH",
		"YE", "ZM", "ZW", "XX",
	}
)
