package event

import (
	"time"

	"github.com/prismelabs/analytics/pkg/uri"
)

// OutboundLinkClick define a click on an outbound link event.
type OutboundLinkClick struct {
	Timestamp time.Time `json:"timestamp"`
	PageUri   uri.Uri   `json:"page_uri"`
	Session   Session   `json:"session"`
	Link      uri.Uri   `json:"link"`
}
