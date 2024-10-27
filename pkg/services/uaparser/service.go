package uaparser

import (
	"strconv"

	"github.com/prismelabs/analytics/pkg/embedded"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/ua-parser/uap-go/uaparser"
)

// Service define a user agent parser service.
type Service interface {
	ParseUserAgent(string) Client
}

// ProvideService is a wire provider for User Agent parser service.
func ProvideService(
	logger zerolog.Logger,
	promRegistry *prometheus.Registry,
) Service {
	logger = logger.With().
		Str("service", "uaparser").
		Logger()

	parser, err := uaparser.NewFromBytes(embedded.UapRegexes)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load user agent parser regexes")
	}

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "uaparser_parse_total",
		Help: "Number of User-Agent parsed",
	}, []string{"browser_family", "operating_system", "device", "is_bot"})
	promRegistry.MustRegister(counter)

	return service{logger, parser, counter}
}

type service struct {
	zerolog.Logger
	parser  *uaparser.Parser
	counter *prometheus.CounterVec
}

// ParseUserAgent implements Service.
func (s service) ParseUserAgent(userAgent string) Client {
	client := s.parser.Parse(userAgent)

	isBot := isBotBrowserFamilyRegex.MatchString(client.UserAgent.Family)

	result := Client{
		BrowserFamily:   client.UserAgent.Family,
		OperatingSystem: client.Os.Family,
		Device:          client.Device.Family,
		IsBot:           isBot,
	}

	if client.Device.Family == "K" {
		client.Device.Family = "Other"
	}

	s.counter.With(prometheus.Labels{
		"browser_family":   result.BrowserFamily,
		"operating_system": result.OperatingSystem,
		"device":           result.Device,
		"is_bot":           strconv.FormatBool(result.IsBot),
	}).Inc()

	s.Logger.Debug().
		Str("user_agent", userAgent).
		Object("client", result).
		Msg("user agent parsed")

	return result
}
