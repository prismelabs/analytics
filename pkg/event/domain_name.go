package event

import (
	"encoding/json"
	"errors"

	"golang.org/x/net/idna"
)

var (
	errValueIsEmpty = errors.New("value is empty")
)

// DomainName define a valid domain name according to RFC 5891. DomainName are
// stored using their ASCII form.
type DomainName struct {
	value string
}

// ParseDomainName parses the given value as a domain name and returns it.
// If the value is considered invalid, an error is returned.
func ParseDomainName(value string) (DomainName, error) {
	if value == "" {
		return DomainName{}, errValueIsEmpty
	}

	domain, err := idna.Lookup.ToASCII(value)
	if err != nil {
		return DomainName{}, err
	}

	return DomainName{domain}, nil
}

// String implements fmt.Stringer.
func (dn DomainName) String() string {
	return dn.value
}

// UnmarshalJSON implements json.Unmarshaler.
func (dn *DomainName) UnmarshalJSON(rawJSON []byte) error {
	var rawDomain string
	err := json.Unmarshal(rawJSON, &rawDomain)
	if err != nil {
		return err
	}

	*dn, err = ParseDomainName(rawDomain)
	if err != nil {
		return err
	}

	return nil
}
