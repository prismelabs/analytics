package event

import (
	"time"

	"github.com/prismelabs/analytics/pkg/uri"
	"github.com/rs/zerolog"
)

var _ zerolog.LogObjectMarshaler = &OutboundLinkClick{}

// FileDownload define a file download.
type FileDownload struct {
	Timestamp time.Time
	PageUri   uri.Uri
	Session   Session
	FileUrl   uri.Uri
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (fd *FileDownload) MarshalZerologObject(e *zerolog.Event) {
	e.
		Time("timestamp", fd.Timestamp).
		Stringer("page_uri", fd.PageUri).
		Object("session", &fd.Session).
		Stringer("file_url", fd.FileUrl)
}
