package main

import (
	"math/rand"
	"strings"
	"time"

	"github.com/negrel/configue"
)

type Config struct {
	TotalEvents       uint64    `json:"total_events"`
	BatchSize         uint64    `json:"batch_size"`
	Domains           []string  `json:"domains"`
	FromDate          time.Time `json:"from_date"`
	CustomEventsRate  float64   `json:"custom_events_rate"`
	BounceRate        float64   `json:"bounce_rate"`
	ExitRate          float64   `json:"exit_rate"`
	MobileRate        float64   `json:"mobile_rate"`
	VisitorIdsRange   uint64    `json:"visitor_ids_range"`
	DirectTrafficRate float64   `json:"direct_traffic_rate"`
}

// RegisterOptions returns a Config parsed from command line.
func (c *Config) RegisterOptions(f *configue.Figue) {
	domains := "localhost,mywebsite.localhost,foo.mywebsite.localhost"
	var extraDomains int
	var extraPaths int

	f.Uint64Var(&c.BatchSize, "batch-size", 40_000, "size of a batch")
	f.Uint64Var(&c.TotalEvents, "total-events", 4_000_000, "number of events to generate")
	f.StringVar(&domains, "domains", domains, "comma separated extra list of domains with events")
	f.IntVar(&extraDomains, "extra-domains", 10, "number of random domains generated added to the domains list")
	f.IntVar(&extraPaths, "extra-paths", 10, "number of random paths generated added to the paths list")
	f.Float64Var(&c.CustomEventsRate, "custom-events-rate", 0.3, "custom events rate per viewed page")
	f.Float64Var(&c.BounceRate, "bounce-rate", 0.56, "bounce rate")
	f.Float64Var(&c.ExitRate, "exit-rate", 0.3, "exit rate when no bounce")
	f.Float64Var(&c.MobileRate, "mobile-rate", 0.3, "mobile client rate")
	f.Uint64Var(&c.VisitorIdsRange, "visitor-ids", 40_000, "range of visitor ids")
	f.Float64Var(&c.DirectTrafficRate, "direct-rate", 0.5, "direct traffic rate against external traffic")

	c.FromDate = time.Now().AddDate(0, -6, 0)

	// Generate extra domains.
	for i := 0; i < extraDomains; i++ {
		c.Domains = append(c.Domains, randomString(alpha, 1)+randomString(alphaNum, rand.Intn(8))+randomItem([]string{".com", ".fr", ".eu", ".io", ".sh"}))
	}

	// Generate extra paths
	for i := 0; i < 1000; i++ {
		part := 1 + rand.Intn(8)
		var path []string
		for j := 0; j < part; j++ {
			path = append(path, "/"+randomString(alphaNum, 1+rand.Intn(8)))
		}

		pathnamesList = append(pathnamesList, strings.Join(path, ""))
	}
}

func (c Config) BatchCount() uint64 {
	return c.TotalEvents / c.BatchSize
}
