package main

import (
	"flag"
	"math"
	"strings"
	"time"
)

// addevents scenarion configuration.
type Config struct {
	BatchSize         int
	BatchCount        int
	Domains           []string
	Paths             []string
	EventType         string
	FromDate          time.Time
	BounceRate        float64
	ExitRate          float64
	VisitorIdsRange   uint64
	DirectTrafficRate float64
}

// ProvideConfig is a wire provider for Config.
func ProvideConfig() Config {
	cfg := Config{}

	domains := "localhost,mywebsite.localhost,foo.mywebsite.localhost"

	flag.IntVar(&cfg.BatchSize, "batch-size", 40_000, "size of a single batch")
	flag.IntVar(&cfg.BatchCount, "batch-count", 1000, "number of batch to send")
	flag.StringVar(&domains, "domains", domains, "comma separated list of domains to target")
	flag.StringVar(&cfg.EventType, "event-type", "pageview", "event type to send (pageview, custom)")
	flag.Float64Var(&cfg.BounceRate, "bounce-rate", 0.56, "bounce rate")
	flag.Float64Var(&cfg.ExitRate, "exit-rate", 0.3, "exit rate when no bounce")
	flag.Uint64Var(&cfg.VisitorIdsRange, "visitor-ids", math.MaxUint64, "range of visitor ids")
	flag.Float64Var(&cfg.DirectTrafficRate, "direct-rate", 1, "direct traffic rate against external traffic")
	flag.Parse()

	cfg.Domains = strings.Split(domains, ",")
	cfg.FromDate = time.Now().AddDate(0, -6, 0)

	return cfg
}
