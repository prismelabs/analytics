package uaparser

import "github.com/rs/zerolog"

// Client define client information derived from user agent.
type Client struct {
	BrowserFamily   string
	OperatingSystem string
	Device          string
	IsBot           bool
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (c Client) MarshalZerologObject(e *zerolog.Event) {
	e.Str("browser_family", c.BrowserFamily).
		Str("operating_system", c.OperatingSystem).
		Str("device", c.Device).
		Bool("is_bot", c.IsBot)
}
