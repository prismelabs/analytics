package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prismelabs/analytics/pkg/uri"
)

// Session holds data about visitor's/user's session/visit.
type Session struct {
	// This data struct must not contains data changing over pageviews and custom events
	// except for PageviewCount (version) field.

	PageUri       uri.Uri                  `json:"page_uri"`
	ReferrerUri   ReferrerUri              `json:"referrer_uri"`
	Client        uaparser.Client          `json:"client"`
	CountryCode   ipgeolocator.CountryCode `json:"country_code"`
	VisitorId     string                   `json:"visitor_id"`
	SessionUuid   uuid.UUID                `json:"session_uuid"`
	Utm           UtmParams                `json:"utm_params"`
	PageviewCount uint16                   `json:"pageview_count"`
}

// SessionTime returns session creation date time.
func (s *Session) SessionTime() time.Time {
	return time.Unix(s.SessionUuid.Time().UnixTime())
}
