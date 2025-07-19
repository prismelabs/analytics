package event

import (
	"time"

	"github.com/prismelabs/analytics/pkg/uri"
)

// Custom define a user defined event with custom properties.
type Custom struct {
	Timestamp time.Time `json:"timestamp"`
	PageUri   uri.Uri   `json:"page_uri"`
	Session   Session   `json:"session"`
	Name      string    `json:"name"`
	Keys      []string  `json:"keys"`
	Values    []string  `json:"values"`
}
