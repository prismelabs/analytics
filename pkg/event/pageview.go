package event

import (
	"time"

	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/rs/zerolog"
)

// PageView define a page view event.
type PageView struct {
	Session    Session
	Timestamp  time.Time
	PageUri    uri.Uri
	TimeOnPage time.Duration
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (pv *PageView) MarshalZerologObject(e *zerolog.Event) {
	e.
		Object("session", &pv.Session).
		Time("timestamp", pv.Timestamp).
		Stringer("page_uri", pv.PageUri).
		Dur("time_on_page", pv.TimeOnPage)
}
