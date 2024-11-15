package uaparser

import (
	"fmt"
	"io"
	"strconv"
	"testing"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	type testCase struct {
		expectedClient Client
		userAgent      string
	}

	testCases := []testCase{
		// user agents copied from https://www.useragents.me/.
		{
			expectedClient: Client{
				BrowserFamily:   "Chrome",
				OperatingSystem: "Windows",
				Device:          "Other",
			},
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.3",
		},
		{
			expectedClient: Client{
				BrowserFamily:   "Safari",
				OperatingSystem: "macOS",
				Device:          "Mac",
			},
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.1",
		},
		{
			expectedClient: Client{
				BrowserFamily:   "Chrome",
				OperatingSystem: "macOS",
				Device:          "Mac",
			},
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.3",
		},
		{
			expectedClient: Client{
				BrowserFamily:   "Other",
				OperatingSystem: "Windows",
				Device:          "Other",
			},
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.",
		},
		{
			expectedClient: Client{
				BrowserFamily:   "Edge",
				OperatingSystem: "Windows",
				Device:          "Other",
			},
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.",
		},
		{
			expectedClient: Client{
				BrowserFamily:   "Chrome",
				OperatingSystem: "Android",
				Device:          "Other",
			},
			userAgent: "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.3",
		},
		{
			expectedClient: Client{
				BrowserFamily:   "Safari",
				OperatingSystem: "iOS",
				Device:          "iPhone",
			},
			userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.",
		},
		{
			expectedClient: Client{
				BrowserFamily:   "bot",
				OperatingSystem: "Other",
				Device:          "Spider",
				IsBot:           true,
			},
			userAgent: "Applebot",
		},
		{
			expectedClient: Client{
				BrowserFamily:   "bot",
				OperatingSystem: "Other",
				Device:          "Spider",
				IsBot:           true,
			},
			userAgent: "AdsBot-Google",
		},
	}

	logger := log.NewLogger("test_logger_1", io.Discard, false)
	service := ProvideService(logger, prometheus.NewRegistry())
	for _, tcase := range testCases {
		testName := fmt.Sprintf("%v/%v/%v", tcase.expectedClient.BrowserFamily, tcase.expectedClient.OperatingSystem, tcase.expectedClient.Device)

		t.Run(testName, func(t *testing.T) {
			cli := service.ParseUserAgent(tcase.userAgent)
			require.Equal(t, tcase.expectedClient, cli)
		})
	}
}

func TestIntegServiceMetrics(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	logger := log.NewLogger("test_logger_1", io.Discard, false)
	promRegistry := prometheus.NewRegistry()
	service := ProvideService(logger, promRegistry)

	client := service.ParseUserAgent("AdsBot-Google")

	require.Equal(t, float64(1), testutils.CounterValue(t, promRegistry, "uaparser_parse_total", prometheus.Labels{
		"browser_family":   client.BrowserFamily,
		"operating_system": client.OperatingSystem,
		"device":           client.Device,
		"is_bot":           strconv.FormatBool(client.IsBot),
	}))
}
