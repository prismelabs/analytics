package event

import (
	"time"

	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/rs/zerolog"
)

var _ zerolog.LogObjectMarshaler = &Custom{}

// Custom define a user defined event with custom properties.
type Custom struct {
	Timestamp   time.Time
	PageUri     Uri
	ReferrerUri ReferrerUri
	Client      uaparser.Client
	CountryCode ipgeolocator.CountryCode
	VisitorId   string
	Name        string
	Keys        []string
	Values      []string
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (c *Custom) MarshalZerologObject(e *zerolog.Event) {
	e.
		Time("timestamp", c.Timestamp).
		Stringer("page_uri", &c.PageUri).
		Stringer("referrer_uri", &c.ReferrerUri).
		Object("client", c.Client).
		Stringer("country_code", c.CountryCode).
		Str("visitor_id", c.VisitorId).
		Str("name", c.Name).
		Strs("keys", c.Keys).
		Strs("values", c.Values)
}
