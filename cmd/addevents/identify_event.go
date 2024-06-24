package main

import (
	"time"

	"github.com/google/uuid"
)

type IdentifyEvent struct {
	Timestamp   time.Time
	SessionUuid uuid.UUID
	VisitorId   string

	// Properties that are set only once (on first identify event).
	InitialKeys   []string
	InitialValues []string

	// Properties that can be overriden.
	Keys   []string
	Values []string
}

// Row encode session as a flat []any slice ready to be sent to ClickHouse.
func (ie IdentifyEvent) Row() (record []any) {
	record = append(record,
		ie.Timestamp,
		ie.VisitorId,
		ie.SessionUuid,
		ie.InitialKeys,
		ie.InitialValues,
		ie.Keys,
		ie.Values,
	)

	return
}
