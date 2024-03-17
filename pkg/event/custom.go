package event

import (
	"encoding/json"
	"errors"
	"regexp"

	"github.com/rs/zerolog"
)

var (
	customNameRegex           = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)
	ErrInvalidCustomEventName = errors.New("invalid custom event name")
	ErrInvalidJson            = errors.New("invalid json")
)

// Custom define a custom events.
type Custom struct {
	domainName DomainName
	name       string
	json       []byte
}

// NewCustom returns a new custom event.
func NewCustom(domainName DomainName, name string, rawJson []byte) (Custom, error) {
	if name == "" {
		return Custom{}, ErrInvalidCustomEventName
	}

	if !json.Valid(rawJson) {
		return Custom{}, ErrInvalidJson
	}

	return Custom{domainName, name, rawJson}, nil
}

// DomainName define domain name associated to event.
func (c Custom) DomainName() DomainName {
	return c.domainName
}

// Name define custom event name.
func (c Custom) Name() string {
	return c.name
}

// Properties defines custom events properties as bytes slice containing well
// formed JSON.
func (c Custom) Properties() []byte {
	return c.json
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (c Custom) MarshalZerologObject(e *zerolog.Event) {
	e.Str("name", c.name).RawJSON("event", c.json)
}
