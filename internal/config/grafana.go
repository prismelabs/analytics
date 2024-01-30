package config

import "github.com/prismelabs/prismeanalytics/internal/secret"

// Grafana related options.
type Grafana struct {
	Url      string
	User     secret.Secret[string]
	Password secret.Secret[string]
}

// GrafanaFromEnv loads grafana related options from environment variables.
// This function panics if required environment variables are missing.
func GrafanaFromEnv() Grafana {
	return Grafana{
		Url:      MustGetEnv("PRISME_GRAFANA_URL"),
		User:     secret.New(MustGetEnv("PRISME_GRAFANA_USER")),
		Password: secret.New(MustGetEnv("PRISME_GRAFANA_PASSWORD")),
	}
}
