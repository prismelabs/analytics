package main

import (
	"fmt"
	"os"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	logger := log.New("ip2country", os.Stderr, false)
	promRegistry := prometheus.NewRegistry()
	ipgeo := ipgeolocator.NewMmdbService(logger, promRegistry)

	for _, arg := range os.Args[1:] {
		fmt.Println(ipgeo.FindCountryCodeForIP(arg).String())
	}
}
