package wired

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// ProvidePrometheusRegistry is a wire provider for a prometheus registry.
func ProvidePrometheusRegistry() *prometheus.Registry {
	registry := prometheus.NewRegistry()

	// Collectors of default prometheus registry.
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())

	return registry
}
