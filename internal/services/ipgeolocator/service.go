package ipgeolocator

// Service define a service responsible of geolocating IP addresses.
type Service interface {
	FindCountryCodeForIP(ip string) CountryCode
}
