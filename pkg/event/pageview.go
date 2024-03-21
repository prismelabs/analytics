package event

import (
	"time"

	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/rs/zerolog"
)

var _ zerolog.LogObjectMarshaler = &PageView{}

// PageView define a page view event.
type PageView struct {
	Timestamp   time.Time
	PageUri     Uri
	ReferrerUri ReferrerUri
	Client      uaparser.Client
	CountryCode ipgeolocator.CountryCode
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (pv *PageView) MarshalZerologObject(e *zerolog.Event) {
	e.
		Time("timestamp", pv.Timestamp).
		Stringer("page_uri", &pv.PageUri).
		Stringer("referrer_uri", &pv.ReferrerUri).
		Object("client", pv.Client).
		Stringer("country_code", pv.CountryCode)
}
