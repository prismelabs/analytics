package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("User agents list file path is missing")
		os.Exit(1)
	}

	logger := log.New("uaparser", os.Stderr, false)
	registry := prometheus.NewRegistry()
	uaParser := uaparser.NewService(logger, registry)

	userAgents, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	if !json.Valid(userAgents) {
		logger.Error("json is invalid, results may be incoherent")
		return
	}

	var userAgentsList []struct {
		UserAgent string `json:"ua"`
	}
	err = json.Unmarshal(userAgents, &userAgentsList)
	if err != nil {
		panic(err)
	}

	var clients []uaparser.Client
	for _, item := range userAgentsList {
		client := uaParser.ParseUserAgent(item.UserAgent)
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
