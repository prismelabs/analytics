package ipgeolocator

import "encoding/json"

// CountryCode define a ISO 3166-1 alpha-2 country code.
type CountryCode struct {
	value string
}

// String implements fmt.Stringer.
func (cc CountryCode) String() string {
	return cc.value
}

// MarshalJSON implements json.Marshaler.
func (cc CountryCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(cc.String())
}
