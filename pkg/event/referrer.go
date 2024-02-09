package event

import (
	"fmt"
	"net/url"

	"golang.org/x/net/idna"
)

// ReferrerDomain define an HTTP referral. A referral is either direct (empty string)
// or a valid absolute URL from which domain is extracted.
type ReferrerDomain struct {
	value string
}

// ParseReferrerDomain parses the given value as a referrer and returns it.
// An error is returned if the value is not a valid referrer.
func ParseReferrerDomain(value string) (ReferrerDomain, error) {
	// Direct source.
	if value == "" {
		return ReferrerDomain{"direct"}, nil
	}

	u, err := url.ParseRequestURI(value)
	if err != nil {
		return ReferrerDomain{}, fmt.Errorf("invalid referrer: %w", err)
	}

	source, err := idna.Lookup.ToASCII(u.Hostname())
	if err != nil {
		return ReferrerDomain{}, err
	}

	return ReferrerDomain{source}, nil
}

// String implements fmt.Stringer.
func (s ReferrerDomain) String() string {
	return s.value
}
