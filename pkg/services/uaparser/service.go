package uaparser

import (
	"strings"

	"github.com/rs/zerolog"
	"github.com/ua-parser/uap-go/uaparser"
)

// Service define a user agent parser service.
type Service interface {
	ParseUserAgent(string) Client
}

// ProvideService is a wire provider for User Agent parser service.
func ProvideService(logger zerolog.Logger) Service {
	logger = logger.With().
		Str("service", "uaparser").
		Logger()

	parser := uaparser.NewFromSaved()

	return service{logger, parser}
}

type service struct {
	zerolog.Logger
	parser *uaparser.Parser
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

	s.Logger.Debug().
		Str("user_agent", userAgent).
		Object("client", result).
		Msg("user agent parsed")

	return result
}
