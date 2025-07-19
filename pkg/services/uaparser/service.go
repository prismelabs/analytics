package uaparser

import (
	"strconv"

	"github.com/prismelabs/analytics/pkg/embedded"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/ua-parser/uap-go/uaparser"
)

// Service define a user agent parser service.
type Service interface {
	ParseUserAgent(string) Client
}

// NewService returns a new User Agent parser service.
func NewService(
	logger log.Logger,
	promRegistry *prometheus.Registry,
) Service {
	logger = logger.With(
		"service", "uaparser",
	)

	parser, err := uaparser.NewFromBytes(embedded.UapRegexes)
	if err != nil {
		logger.Fatal("failed to load user agent parser regexes", err)
	}

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "uaparser_parse_total",
		Help: "Number of User-Agent parsed",
	}, []string{"browser_family", "operating_system", "device", "is_bot"})
	promRegistry.MustRegister(counter)

	return service{logger, parser, counter}
}

type service struct {
	logger  log.Logger
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

	// https://www.youtube.com/watch?v=ftDVCo8SFD4
	if result.Device == "K" {
		result.Device = "Other"
	}

	s.counter.With(prometheus.Labels{
		"browser_family":   result.BrowserFamily,
		"operating_system": result.OperatingSystem,
		"device":           result.Device,
		"is_bot":           strconv.FormatBool(result.IsBot),
	}).Inc()

	s.logger.Debug(
		"user agent parsed",
		"user_agent", userAgent,
		"client", result,
	)

	return result
}
