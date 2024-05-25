package event

import (
	"math/big"
	"time"

	"github.com/rs/zerolog"
)

var _ zerolog.LogObjectMarshaler = &PageView{}

// PageView define a page view event.
type PageView struct {
	Timestamp time.Time
	PageUri   Uri
	VisitorId string
	SessionId *big.Int
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (pv *PageView) MarshalZerologObject(e *zerolog.Event) {
	e.
		Time("timestamp", pv.Timestamp).
		Stringer("page_uri", &pv.PageUri).
		Str("visitor_id", pv.VisitorId).
		Stringer("session_id", pv.SessionId)
}
