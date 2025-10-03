//go:build test

package ipgeolocator

// NewCountryCode returns a CountryCode containing provided code.
// This function is only available in tests.
func NewCountryCode(code string) CountryCode {
	if len(code) != 2 {
		panic("invalid country code")
	}

	return CountryCode{code}
}
