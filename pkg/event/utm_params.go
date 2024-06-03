package event

import "github.com/rs/zerolog"

// UtmParams holds Urchin Tracking Module (UTM) URL parameters.
// See https://en.wikipedia.org/wiki/UTM_parameters.
type UtmParams struct {
	Source   string
	Medium   string
	Campaign string
	Term     string
	Content  string
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (up *UtmParams) MarshalZerologObject(e *zerolog.Event) {
	e.
		Str("source", up.Source).
		Str("medium", up.Medium).
		Str("campaign", up.Campaign).
		Str("term", up.Term).
		Str("content", up.Content)
}
