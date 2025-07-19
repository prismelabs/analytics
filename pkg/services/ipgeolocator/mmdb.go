package ipgeolocator

import (
	"fmt"
	"net"
	"strings"

	"github.com/oschwald/maxminddb-golang"
	"github.com/prismelabs/analytics/pkg/embedded"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
)

// NewMmdbService returns a new mmdb based ip geolocator service.
func NewMmdbService(
	logger log.Logger,
	promRegistry *prometheus.Registry,
) Service {
	logger = logger.With(
		"service", "ipgeolocator",
		"service_impl", "mmdb",
		"mmdb", "embedded",
	)

	reader, err := maxminddb.FromBytes(embedded.Ip2AsnDb)
	if err != nil {
		logger.Fatal("failed to load maxming GeoLite2 country database", err)
	}

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "ipgeolocator_search_total",
		Help: "Number of IP geolocation",
	}, []string{"country_code", "ip_version"})

	promRegistry.MustRegister(counter)

	return mmdbService{logger, reader, counter}
}

type mmdbService struct {
	logger  log.Logger
	reader  *maxminddb.Reader
	counter *prometheus.CounterVec
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

	ipVersion := "6"

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

		// Database embedded within repository returns None, Unknown sometime.
		// Official maxmind GeoLite2 database doesn't returns anything.
		// If ISO code is not valid 2 letter code, default to XX.
		if len(record.Country.ISOCode) != 2 {
			record.Country.ISOCode = "XX"
		}

		// IPv4 address
		if ipAddr.To4() != nil {
			ipVersion = "4"
		}

		result.value = record.Country.ISOCode
		break
	}

	// Increment metric.
	ms.counter.With(prometheus.Labels{
		"country_code": result.value,
		"ip_version":   ipVersion,
	}).Inc()

	ms.logger.Debug(
		"country code found",
		"ip_address", xForwardedFor,
		"country_code", result.value,
	)
	return result
}
