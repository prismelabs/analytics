package event

import (
	"time"

	"github.com/prismelabs/analytics/pkg/uri"
)

// FileDownload define a file download.
type FileDownload struct {
	Timestamp time.Time `json:"timestamp"`
	PageUri   uri.Uri   `json:"page_uri"`
	Session   Session   `json:"session"`
	FileUrl   uri.Uri   `json:"file_url"`
}
