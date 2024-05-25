package event

import (
	"math/big"
	"time"

	"github.com/rs/zerolog"
)

var _ zerolog.LogObjectMarshaler = &Custom{}

// Custom define a user defined event with custom properties.
type Custom struct {
	Timestamp time.Time
	PageUri   Uri
	VisitorId string
	SessionId *big.Int
	Name      string
	Keys      []string
	Values    []string
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (c *Custom) MarshalZerologObject(e *zerolog.Event) {
	e.
		Time("timestamp", c.Timestamp).
		Stringer("page_uri", &c.PageUri).
		Str("visitor_id", c.VisitorId).
		Stringer("session_id", c.SessionId).
		Str("name", c.Name).
		Strs("keys", c.Keys).
		Strs("values", c.Values)
}
