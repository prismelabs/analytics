package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("User agents list file path is missing")
		os.Exit(1)
	}

	path := ""
	if len(os.Args) == 3 {
		path = os.Args[2]
	}

	logger := log.NewLogger("uaparser", os.Stderr, false)
	registry := prometheus.NewRegistry()
	uaParser := uaparser.ProvideService(logger, registry)

	userAgents, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	if !gjson.ValidBytes(userAgents) {
		logger.Warn().Msg("json is invalid, results may be incoherent")
	}

	userAgentsList := gjson.Parse(string(userAgents)).Array()

	var clients []uaparser.Client
	for _, item := range userAgentsList {
		ua := item.String()
		if path != "" {
			ua = item.Get(path).String()
		}
		client := uaParser.ParseUserAgent(ua)
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
