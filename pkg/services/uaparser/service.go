package uaparser

import "github.com/ua-parser/uap-go/uaparser"

type Service interface {
	ParseUserAgent(string) Client
}

// ProvideService is a wire provider for User Agent parser service.
func ProvideService() Service {
	parser := uaparser.NewFromSaved()
	return service{parser}
}

type service struct {
	parser *uaparser.Parser
}

// ParseUserAgent implements Service.
func (s service) ParseUserAgent(userAgent string) Client {
	client := s.parser.Parse(userAgent)

	return Client{
		BrowserFamily:   client.UserAgent.Family,
		OperatingSystem: client.Os.Family,
		Device:          client.Device.Family,
	}
}
