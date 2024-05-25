package event

import (
	"math/big"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/rs/zerolog"
)

// Session define a session event.
type Session struct {
	SessionUuid uuid.UUID
	PageView
	ReferrerUri ReferrerUri
	Client      uaparser.Client
	CountryCode ipgeolocator.CountryCode
}

// SessionId returns session id.
func (s *Session) SessionId() *big.Int {
	if s.PageView.SessionId != nil {
		return s.PageView.SessionId
	}

	uuidBytes := ([16]byte(s.SessionUuid))
	return big.NewInt(0).SetBytes(uuidBytes[:])
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (s *Session) MarshalZerologObject(e *zerolog.Event) {
	e.
		Object("pageview", &s.PageView).
		Stringer("referrer_uri", &s.ReferrerUri).
		Str("operating_system", s.Client.OperatingSystem).
		Str("browser_family", s.Client.BrowserFamily).
		Str("device", s.Client.Device).
		Bool("is_bot", s.Client.IsBot).
		Stringer("session_uuid", s.SessionUuid)
}
