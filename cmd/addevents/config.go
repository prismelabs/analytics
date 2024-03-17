package main

import (
	"flag"
	"strings"
	"time"
)

type Config struct {
	BatchSize  int
	BatchCount int
	Domains    []string
	Paths      []string
	EventType  string
	FromDate   time.Time
}

func ProvideConfig() Config {
	cfg := Config{}

	domains := "localhost,mywebsite.localhost,foo.mywebsite.localhost"

	flag.IntVar(&cfg.BatchSize, "batchsize", 40_000, "size of a single batch")
	flag.IntVar(&cfg.BatchCount, "batchcount", 1_000, "number of batch to send")
	flag.StringVar(&domains, "domains", domains, "comma separated list of domains to target")
	flag.StringVar(&cfg.EventType, "eventtype", "pageview", "event type to send (pageview, custom)")
	flag.Parse()

	cfg.Domains = strings.Split(domains, ",")
	cfg.FromDate = time.Now().AddDate(0, -6, 0)

	return cfg
}
