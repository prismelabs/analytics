package event

import (
	"time"

	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/rs/zerolog"
)

var _ zerolog.LogObjectMarshaler = &Custom{}

// Custom define a user defined event with custom properties.
type Custom struct {
	Timestamp time.Time
	PageUri   uri.Uri
	Session   Session
	Name      string
	Keys      []string
	Values    []string
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (c *Custom) MarshalZerologObject(e *zerolog.Event) {
	e.
		Time("timestamp", c.Timestamp).
		Stringer("page_uri", c.PageUri).
		Object("session", &c.Session).
		Str("name", c.Name).
		Strs("keys", c.Keys).
		Strs("values", c.Values)
}
