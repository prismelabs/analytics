package main

import (
	_ "embed"
	"encoding/json"

	"github.com/prismelabs/analytics/pkg/services/uaparser"
)

//go:embed rand_data/desktop_user_agents.json
var desktopUserAgents []byte

//go:embed rand_data/mobile_user_agents.json
var mobileUserAgents []byte

var desktopClients []uaparser.Client
var mobileClients []uaparser.Client

func init() {
	err := json.Unmarshal(desktopUserAgents, &desktopClients)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(mobileUserAgents, &mobileClients)
	if err != nil {
		panic(err)
	}
}
