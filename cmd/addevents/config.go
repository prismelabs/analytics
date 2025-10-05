package main

import (
	"runtime"

	"github.com/negrel/configue"
)

type Config struct {
	BounceRate          float64 `json:"bounce_rate"`
	CustomEventsRate    float64 `json:"custom_events_rate"`
	DirectTrafficRate   float64 `json:"direct_traffic_rate"`
	MobileRate          float64 `json:"mobile_rate"`
	TotalSessions       uint64  `json:"total_sessions"`
	PageViewsPerSession float64 `json:"view_per_session"`
	Workers             uint64  `json:"workers"`
}

// RegisterOptions registers options in provided Figue.
func (c *Config) RegisterOptions(f *configue.Figue) {
	f.Float64Var(&c.BounceRate, "bounce.rate", 0.56, "bounce rate")
	f.Float64Var(&c.CustomEventsRate, "custom.events-rate", 0.3, "custom events rate per viewed page")
	f.Float64Var(&c.DirectTrafficRate, "direct.rate", 0.5, "direct traffic rate against external traffic")
	f.Float64Var(&c.MobileRate, "mobile.rate", 0.3, "mobile client rate")
	f.Uint64Var(&c.TotalSessions, "total.sessions", 4_000_000, "number of sessions")
	f.Float64Var(&c.PageViewsPerSession, "views.per.session", 3.5, "average number of page view per session")
	f.Uint64Var(&c.Workers, "workers", min(1, uint64(runtime.NumCPU()/2)), "number of worker")
}
