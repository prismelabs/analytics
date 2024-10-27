package event

import (
	"time"

	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/rs/zerolog"
)

var _ zerolog.LogObjectMarshaler = &Click{}

// Click define a click event.
type Click struct {
	Timestamp time.Time
	PageUri   uri.Uri
	Session   Session
	Tag       string
	Id        string
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (c *Click) MarshalZerologObject(e *zerolog.Event) {
	e.
		Time("timestamp", c.Timestamp).
		Stringer("page_uri", c.PageUri).
		Object("session", &c.Session).
		Str("tag", c.Tag).
		Str("id", c.Id)
}
