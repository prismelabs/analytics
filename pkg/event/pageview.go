package event

import (
	"net/url"
	"path"
	"time"

	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/rs/zerolog"
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
	domainName DomainName,
	cli uaparser.Client,
	pageReferrer string,
	countryCode ipgeolocator.CountryCode,
) (PageView, error) {
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
		DomainName:     domainName,
		PathName:       path.Clean(pageviewPath),
		Client:         cli,
		ReferrerDomain: referrerDomain,
		CountryCode:    countryCode,
	}, nil
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (pv PageView) MarshalZerologObject(e *zerolog.Event) {
	e.Time("timestamp", pv.Timestamp).
		Stringer("domain_name", pv.DomainName).
		Str("path", pv.PathName).
		Object("client", pv.Client).
		Stringer("referrer_domain", pv.ReferrerDomain).
		Stringer("country_code", pv.CountryCode)
}
