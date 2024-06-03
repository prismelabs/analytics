package event

import (
	"time"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/rs/zerolog"
)

// Session holds data about visitor's/user's session/visit.
type Session struct {
	PageUri     *Uri
	ReferrerUri *ReferrerUri
	Client      uaparser.Client
	CountryCode ipgeolocator.CountryCode
	VisitorId   string
	SessionUuid uuid.UUID
	Utm         UtmParams
	Pageviews   uint16
}

// Version returns session version number.
// Higher version number means more recent session.
func (s *Session) Version() uint16 {
	return s.Pageviews
}

// SessionTime returns session creation date time.
func (s *Session) SessionTime() time.Time {
	return time.Unix(s.SessionUuid.Time().UnixTime())
}

func (s *Session) MarshalZerologObject(e *zerolog.Event) {
	e.
		Stringer("page_uri", s.ReferrerUri).
		Stringer("referrer_uri", s.ReferrerUri).
		Object("client", s.Client).
		Stringer("country_code", s.CountryCode).
		Str("visitor_id", s.VisitorId).
		Stringer("session_uuid", s.SessionUuid).
		Time("session_time", s.SessionTime()).
		Object("utp_params", &s.Utm).
		Uint16("pageviews", s.Pageviews)
}
