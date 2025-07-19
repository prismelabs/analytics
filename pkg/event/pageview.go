package event

import (
	"time"

	"github.com/prismelabs/analytics/pkg/uri"
)

// PageView define a page view event.
// See https://www.prismeanalytics.com/docs/references/http/#page-view-events
type PageView struct {
	Session   Session   `json:"session"`
	Timestamp time.Time `json:"timestamp"`
	PageUri   uri.Uri   `json:"page_uri"`
	Status    uint16    `json:"status"`
}
