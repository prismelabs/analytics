package main

import (
	"flag"
	"math"
	"math/rand"
	"strings"
	"time"
)

type Config struct {
	TotalEvents       uint64
	BatchSize         uint64
	Domains           []string
	FromDate          time.Time
	CustomEventsRate  float64
	BounceRate        float64
	ExitRate          float64
	MobileRate        float64
	VisitorIdsRange   uint64
	DirectTrafficRate float64
}

// ProvideConfig is a wire provider for Config.
func ProvideConfig() Config {
	cfg := Config{}

	domains := "localhost,mywebsite.localhost,foo.mywebsite.localhost"
	var extraDomains int
	var extraPaths int

	flag.Uint64Var(&cfg.BatchSize, "batch-size", 40_000, "size of a batch")
	flag.Uint64Var(&cfg.TotalEvents, "total-events", 40_000_000, "number of events to generate")
	flag.StringVar(&domains, "domains", domains, "comma separated extra list of domains with events")
	flag.IntVar(&extraDomains, "extra-domains", 10, "number of random domains generated added to the domains list")
	flag.IntVar(&extraPaths, "extra-paths", 10, "number of random paths generated added to the paths list")
	flag.Float64Var(&cfg.CustomEventsRate, "custom-events-rate", 0.3, "custom events rate per viewed page")
	flag.Float64Var(&cfg.BounceRate, "bounce-rate", 0.56, "bounce rate")
	flag.Float64Var(&cfg.ExitRate, "exit-rate", 0.3, "exit rate when no bounce")
	flag.Float64Var(&cfg.MobileRate, "mobile-rate", 0.3, "mobile client rate")
	flag.Uint64Var(&cfg.VisitorIdsRange, "visitor-ids", math.MaxUint64, "range of visitor ids")
	flag.Float64Var(&cfg.DirectTrafficRate, "direct-rate", 0.5, "direct traffic rate against external traffic")
	flag.Parse()

	cfg.Domains = strings.Split(domains, ",")
	cfg.FromDate = time.Now().AddDate(0, -6, 0)

	// Generate extra domains.
	for i := 0; i < extraDomains; i++ {
		cfg.Domains = append(cfg.Domains, randomString(alpha, 1)+randomString(alphaNum, rand.Intn(8))+randomItem([]string{".com", ".fr", ".eu", ".io", ".sh"}))
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

	return cfg
}

func (c Config) BatchCount() uint64 {
	return c.TotalEvents / c.BatchSize
}
