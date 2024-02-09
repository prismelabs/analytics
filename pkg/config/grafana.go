package config

import "github.com/prismelabs/prismeanalytics/pkg/secret"

// Grafana related options.
type Grafana struct {
	Url      string
	User     secret.Secret[string]
	Password secret.Secret[string]
	OrgId    int64
}

// GrafanaFromEnv loads grafana related options from environment variables.
// This function panics if required environment variables are missing.
func GrafanaFromEnv() Grafana {
	return Grafana{
		Url:      MustGetEnv("PRISME_GRAFANA_URL"),
		User:     secret.New(MustGetEnv("PRISME_GRAFANA_USER")),
		Password: secret.New(MustGetEnv("PRISME_GRAFANA_PASSWORD")),
		OrgId:    ParseIntEnvOrDefault("PRISME_GRAFANA_ORG_ID", 1, 64),
	}
}
