package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prismelabs/analytics/pkg/wired"
)

// JSON from https://useragents.me
type userAgentsMe struct {
	Ua string `json:"ua"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("User agents list file path is missing")
		os.Exit(1)
	}

	logger := log.NewLogger("uaparser", os.Stderr, false)
	registry := wired.ProvidePrometheusRegistry()
	uaParser := uaparser.ProvideService(logger, registry)

	userAgents, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	var userAgentsList []userAgentsMe
	err = json.Unmarshal(userAgents, &userAgentsList)
	if err != nil {
		panic(err)
	}

	var clients []uaparser.Client
	for _, item := range userAgentsList {
		client := uaParser.ParseUserAgent(item.Ua)
		clients = append(clients, client)
	}

	clientsJson, err := json.Marshal(clients)
	if err != nil {
		panic(err)
	}

	_, err = os.Stdout.Write(clientsJson)
	if err != nil {
		panic(err)
	}
}
