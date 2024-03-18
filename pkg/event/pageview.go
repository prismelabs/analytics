package event

import (
	"time"

	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
)

// PageView define a page view event.
type PageView struct {
	Timestamp   time.Time
	PageUri     Uri
	ReferrerUri ReferrerUri
	Client      uaparser.Client
	CountryCode ipgeolocator.CountryCode
}
