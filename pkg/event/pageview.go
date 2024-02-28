package event

import (
	"net/url"
	"path"
	"time"

	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
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
func NewPageView(
	pvUrl *url.URL,
	cli uaparser.Client,
	pageReferrer string,
	countryCode ipgeolocator.CountryCode,
) (PageView, error) {
	domain, err := ParseDomainName(pvUrl.Hostname())
	if err != nil {
		return PageView{}, err
	}

	referrerDomain, err := ParseReferrerDomain(pageReferrer)
	if err != nil {
		return PageView{}, err
	}

	pageviewPath := pvUrl.Path
	if pageviewPath == "" {
		pageviewPath = "/"
	}

	return PageView{
		Timestamp:      time.Now().UTC(),
		DomainName:     domain,
		PathName:       path.Clean(pageviewPath),
		Client:         cli,
		ReferrerDomain: referrerDomain,
		CountryCode:    countryCode,
	}, nil
}
