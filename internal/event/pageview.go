package event

import (
	"net/url"
	"time"

	"github.com/prismelabs/prismeanalytics/internal/services/uaparser"
)

// OperatingSystem define client operating system.
type OperatingSystem string

// PageView define a page view event.
type PageView struct {
	Timestamp  time.Time
	DomainName DomainName
	PathName   string
	Client     uaparser.Client
}

// NewPageView creates a new PageView event.
func NewPageView(u *url.URL, cli uaparser.Client) (PageView, error) {
	domain, err := ParseDomainName(u.Hostname())
	if err != nil {
		return PageView{}, err
	}

	return PageView{
		Timestamp:  time.Now(),
		DomainName: domain,
		PathName:   u.Path,
		Client:     cli,
	}, nil
}
