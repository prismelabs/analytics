package event

import (
	"time"

	"github.com/rs/zerolog"
)

var _ zerolog.LogObjectMarshaler = &Custom{}

// Custom define a user defined event with custom properties.
type Custom struct {
	Timestamp  time.Time
	Name       string
	PageUri    Uri
	Properties []byte
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (c *Custom) MarshalZerologObject(e *zerolog.Event) {
	e.
		Time("timestamp", c.Timestamp).
		Str("name", c.Name).
		Stringer("page_uri", &c.PageUri).
		Str("properties", string(c.Properties))

}
