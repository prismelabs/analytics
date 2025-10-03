package faker

import (
	_ "embed"
	"encoding/json"
	"io"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prometheus/client_golang/prometheus"
)

//go:embed embedded/desktop_user_agents.json
var desktopUserAgentsJson []byte

//go:embed embedded/mobile_user_agents.json
var mobileUserAgentsJson []byte

var DesktopUserAgents []string
var MobileUserAgents []string

var DesktopClients []uaparser.Client
var MobileClients []uaparser.Client

func init() {
	type UserAgent struct {
		Ua string `json:"ua"`
	}

	logger := log.New("", io.Discard, false)
	uap := uaparser.NewService(logger, prometheus.NewRegistry())

	var userAgents []UserAgent

	err := json.Unmarshal(desktopUserAgentsJson, &userAgents)
	if err != nil {
		panic(err)
	}
	for _, ua := range userAgents {
		DesktopUserAgents = append(DesktopUserAgents, ua.Ua)
		DesktopClients = append(DesktopClients, uap.ParseUserAgent(ua.Ua))
	}

	userAgents = nil
	err = json.Unmarshal(mobileUserAgentsJson, &userAgents)
	if err != nil {
		panic(err)
	}
	for _, ua := range userAgents {
		MobileUserAgents = append(MobileUserAgents, ua.Ua)
		MobileClients = append(MobileClients, uap.ParseUserAgent(ua.Ua))
	}
}
