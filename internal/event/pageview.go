package event

import (
	"net/url"
	"time"

	"github.com/prismelabs/prismeanalytics/internal/services/ipgeolocator"
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
	CountryCode    ipgeolocator.CountryCode
}

// NewPageView creates a new PageView event.
func NewPageView(pvUrl *url.URL,
	cli uaparser.Client,
	pageReferrer string,
	countryCode ipgeolocator.CountryCode) (PageView, error) {
	domain, err := ParseDomainName(pvUrl.Hostname())
	if err != nil {
		return PageView{}, err
	}

	referrerDomain, err := ParseReferrerDomain(pageReferrer)
	if err != nil {
		return PageView{}, err
	}

	path := pvUrl.Path
	if path == "" {
		path = "/"
	}

	return PageView{
		Timestamp:      time.Now(),
		DomainName:     domain,
		PathName:       path,
		Client:         cli,
		ReferrerDomain: referrerDomain,
		CountryCode:    countryCode,
	}, nil
}
