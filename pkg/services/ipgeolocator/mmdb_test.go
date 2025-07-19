package ipgeolocator

import (
	"io"
	"testing"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestMmdbService(t *testing.T) {
	logger := log.New("ipgeolocator_mmdb_service", io.Discard, false)

	t.Run("FindCountryCodeForIP", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			type testCase struct {
				ip          string
				countryCode string
			}

			testCases := []testCase{
				{
					ip:          "127.0.0.1", // Loopback address.
					countryCode: "XX",
				},
				{
					ip:          "127.10.0.2", // Loopback address.
					countryCode: "XX",
				},
				{
					ip:          "10.0.0.0", // Network address.
					countryCode: "XX",
				},
				{
					ip:          "10.1.1.1",
					countryCode: "XX",
				},
				{
					ip:          "172.16.1.10",
					countryCode: "XX",
				},
				{
					ip:          "192.168.1.253",
					countryCode: "XX",
				},
				{
					ip:          "8.8.8.8", // Google public DNS based in US.
					countryCode: "US",
				},
				{
					ip:          "2001:4860:4860::8888", // Google public DNS based in US.
					countryCode: "US",
				},
				{
					ip:          "84.200.69.80", // DNS.watch public DNS based in DE
					countryCode: "DE",
				},
				{
					ip:          "2001:1608:10:25::1c04:b12f", // DNS.watch public DNS based in DE
					countryCode: "DE",
				},
			}

			for _, tcase := range testCases {
				t.Run(tcase.ip+"/"+tcase.countryCode, func(t *testing.T) {
					service := NewMmdbService(logger, prometheus.NewRegistry())

					countryCode := service.FindCountryCodeForIP(tcase.ip)
					require.Equal(t, tcase.countryCode, countryCode.String())
				})
			}
		})

		t.Run("Invalid", func(t *testing.T) {
			type testCase struct {
				ip string
			}
			testCases := []testCase{
				{
					ip: "",
				},
				{
					ip: "not_an_ip",
				},
			}

			for _, tcase := range testCases {
				t.Run(tcase.ip, func(t *testing.T) {
					service := NewMmdbService(logger, prometheus.NewRegistry())

					countryCode := service.FindCountryCodeForIP(tcase.ip)
					require.Equal(t, CountryCode{"XX"}, countryCode)
				})
			}
		})
	})
}

func TestIntegMmdbServiceMetrics(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	logger := log.New("ipgeolocator_mmdb_service", io.Discard, false)
	promRegistry := prometheus.NewRegistry()
	service := NewMmdbService(logger, promRegistry)

	countryCode := service.FindCountryCodeForIP("127.0.0.1")

	require.Equal(t, float64(1), testutils.CounterValue(t, promRegistry, "ipgeolocator_search_total", prometheus.Labels{
		"country_code": countryCode.String(),
		"ip_version":   "4",
	}))
}
