package uaparser

// Client define client information derived from user agent.
type Client struct {
	BrowserFamily   string `json:"browser_family"`
	OperatingSystem string `json:"operating_system"`
	Device          string `json:"device"`
	IsBot           bool   `json:"is_bot"`
}
