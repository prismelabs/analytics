package ipgeolocator

import (
	"fmt"
	"net"
	"strings"

	"github.com/oschwald/maxminddb-golang"
	"github.com/prismelabs/analytics/pkg/embedded"
	"github.com/prismelabs/analytics/pkg/log"
)

// ProvideMmdbService is a wire provider for mmdb based ip geolocator service.
func ProvideMmdbService(logger log.Logger) Service {
	reader, err := maxminddb.FromBytes(embedded.Ip2AsnDb)
	if err != nil {
		logger.Err(err).Msg("failed to load maxming GeoLite2 country database")
	}

	return mmdbService{reader}
}

type mmdbService struct {
	reader *maxminddb.Reader
}

// FindCountryCodeForIP implements Service.
func (ms mmdbService) FindCountryCodeForIP(xForwardedFor string) CountryCode {
	type (
		mmdbRecordCountry struct {
			ISOCode string `maxminddb:"iso_code"`
		}
		mmdbRecord struct {
			Country mmdbRecordCountry `maxminddb:"country"`
		}
	)

	record := mmdbRecord{mmdbRecordCountry{"XX"}}

	// Split has X-Forwarded-For may contains multiple IPs address.
	ips := strings.Split(xForwardedFor, ",")

	// Lookup first valid ip address.
	for _, ip := range ips {
		ipAddr := net.ParseIP(ip)
		if ipAddr == nil {
			continue
		}

		err := ms.reader.Lookup(ipAddr, &record)
		if err != nil {
			panic(fmt.Errorf("failed to lookup ip address in mmdb: %w", err))
		}

		// Database embedded within repository returns None sometime.
		// Official maxmind GeoLite2 database doesn't returns anything.
		if record.Country.ISOCode == "None" {
			record.Country.ISOCode = "XX"
		}

		return CountryCode{record.Country.ISOCode}
	}

	return CountryCode{"XX"}
}
