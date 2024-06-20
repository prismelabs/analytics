package event

import (
	"time"

	"github.com/rs/zerolog"
)

// Identify event adds contextual information.
type Identify struct {
	Timestamp time.Time
	Session   Session

	// Properties that are set only once (on first identify event).
	InitialKeys   []string
	InitialValues []string

	// Properties that can be overriden.
	Keys   []string
	Values []string
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (i *Identify) MarshalZerologObject(e *zerolog.Event) {
	e.
		Time("timestamp", i.Timestamp).
		Object("session", &i.Session).
		Strs("initial_keys", i.InitialKeys).
		Strs("initial_values", i.InitialValues).
		Strs("keys", i.Keys).
		Strs("values", i.Values)
}
