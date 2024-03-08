package ipgeolocator

import (
	"fmt"
	"net"
	"strings"

	"github.com/oschwald/maxminddb-golang"
	"github.com/prismelabs/analytics/pkg/embedded"
	"github.com/rs/zerolog"
)

// ProvideMmdbService is a wire provider for mmdb based ip geolocator service.
func ProvideMmdbService(logger zerolog.Logger) Service {
	logger = logger.With().
		Str("service", "ipgeolocator").
		Str("service_impl", "mmdb").
		Str("mmdb", "embedded").
		Logger()

	reader, err := maxminddb.FromBytes(embedded.Ip2AsnDb)
	if err != nil {
		logger.Panic().Err(err).Msg("failed to load maxming GeoLite2 country database")
	}

	return mmdbService{logger, reader}
}

type mmdbService struct {
	logger zerolog.Logger
	reader *maxminddb.Reader
}

// FindCountryCodeForIP implements Service.
func (ms mmdbService) FindCountryCodeForIP(xForwardedFor string) CountryCode {
	result := CountryCode{"XX"}

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
		if record.Country.ISOCode == "None" || record.Country.ISOCode == "" {
			record.Country.ISOCode = "XX"
		}

		result.value = record.Country.ISOCode
		break
	}

	ms.logger.Debug().
		Str("ip_address", xForwardedFor).
		Str("country_code", result.value).
		Msg("country code found")
	return result
}
