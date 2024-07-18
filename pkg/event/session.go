package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/rs/zerolog"
)

// Session holds data about visitor's/user's session/visit.
type Session struct {
	// This data struct must not contains data changing over pageviews and custom events
	// except for PageviewCount (version) field.

	PageUri       uri.Uri
	ReferrerUri   ReferrerUri
	Client        uaparser.Client
	CountryCode   ipgeolocator.CountryCode
	VisitorId     string
	SessionUuid   uuid.UUID
	Utm           UtmParams
	PageviewCount uint16
}

// SessionTime returns session creation date time.
func (s *Session) SessionTime() time.Time {
	return time.Unix(s.SessionUuid.Time().UnixTime())
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (s *Session) MarshalZerologObject(e *zerolog.Event) {
	e.
		Stringer("page_uri", s.PageUri).
		Stringer("referrer_uri", s.ReferrerUri).
		Object("client", s.Client).
		Stringer("country_code", s.CountryCode).
		Str("visitor_id", s.VisitorId).
		Stringer("session_uuid", s.SessionUuid).
		Time("session_time", s.SessionTime()).
		Object("utp_params", &s.Utm).
		Uint16("pageview_count", s.PageviewCount)
}
