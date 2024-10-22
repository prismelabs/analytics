package config

import (
	"github.com/negrel/secrecy"
)

// Grafana related options.
type Grafana struct {
	Url      string
	User     secrecy.SecretString
	Password secrecy.SecretString
	OrgId    int64
}

// GrafanaFromEnv loads grafana related options from environment variables.
// This function panics if required environment variables are missing.
func GrafanaFromEnv() Grafana {
	return Grafana{
		Url:      MustGetEnv("PRISME_GRAFANA_URL"),
		User:     secrecy.NewSecretString(secrecy.UnsafeStringToBytes(MustGetEnv("PRISME_GRAFANA_USER"))),
		Password: secrecy.NewSecretString(secrecy.UnsafeStringToBytes(MustGetEnv("PRISME_GRAFANA_PASSWORD"))),
		OrgId:    ParseIntEnvOrDefault("PRISME_GRAFANA_ORG_ID", 1, 64),
	}
}
