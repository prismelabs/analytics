package uaparser

import (
	"strconv"
	"strings"

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

	parser := uaparser.NewFromSaved()

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
	result := Client{
		BrowserFamily:   client.UserAgent.Family,
		OperatingSystem: client.Os.Family,
		Device:          client.Device.Family,
		IsBot:           strings.Contains(client.UserAgent.Family, "Bot") || strings.Contains(client.UserAgent.Family, "bot") || strings.Contains(client.Device.Family, "Spider"),
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
