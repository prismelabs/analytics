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
	Timestamp      time.Time
	DomainName     DomainName
	PathName       string
	Client         uaparser.Client
	ReferrerDomain ReferrerDomain
}

// NewPageView creates a new PageView event.
func NewPageView(pvUrl *url.URL, cli uaparser.Client, pageReferrer string) (PageView, error) {
	domain, err := ParseDomainName(pvUrl.Hostname())
	if err != nil {
		return PageView{}, err
	}

	referrerDomain, err := ParseReferrerDomain(pageReferrer)
	if err != nil {
		return PageView{}, err
	}

	return PageView{
		Timestamp:      time.Now(),
		DomainName:     domain,
		PathName:       pvUrl.Path,
		Client:         cli,
		ReferrerDomain: referrerDomain,
	}, nil
}
