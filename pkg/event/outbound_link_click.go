package event

import (
	"time"

	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/rs/zerolog"
)

var _ zerolog.LogObjectMarshaler = &OutboundLinkClick{}

// OutboundLinkClick define a click on an outbound link event.
type OutboundLinkClick struct {
	Timestamp time.Time
	PageUri   uri.Uri
	Session   Session
	Link      uri.Uri
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (olc *OutboundLinkClick) MarshalZerologObject(e *zerolog.Event) {
	e.
		Time("timestamp", olc.Timestamp).
		Stringer("page_uri", olc.PageUri).
		Object("session", &olc.Session).
		Stringer("link", olc.Link)
}
