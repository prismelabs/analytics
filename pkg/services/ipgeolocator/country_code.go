package ipgeolocator

// CountryCode define a ISO 3166-1 alpha-2 country code.
type CountryCode struct {
	value string
}

// String implements fmt.Stringer.
func (cc CountryCode) String() string {
	return cc.value
}
