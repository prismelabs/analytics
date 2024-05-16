package main

import (
	"github.com/rs/zerolog"
	"sync/atomic"
)

type Metrics struct {
	events  atomic.Uint64
	bounces atomic.Uint64
	visits  atomic.Uint64
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (m *Metrics) MarshalZerologObject(e *zerolog.Event) {
	e.Uint64("events", m.events.Load()).
		Uint64("bounces", m.bounces.Load()).
		Uint64("visits", m.visits.Load())
}
