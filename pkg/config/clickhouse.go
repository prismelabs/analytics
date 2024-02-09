package config

import (
	"github.com/prismelabs/prismeanalytics/pkg/secret"
)

// Clickhouse connection options.
type Clickhouse struct {
	TlsEnabled bool
	HostPort   string
	Database   string
	User       secret.Secret[string]
	Password   secret.Secret[string]
}

// ClickhouseFromEnv loads clickhouse config from environment variables.
// This function panics if required environment variables are missing.
func ClickhouseFromEnv() Clickhouse {
	return Clickhouse{
		TlsEnabled: GetEnvOrDefault("PRISME_CLICKHOUSE_TLS", "false") != "false",
		HostPort:   MustGetEnv("PRISME_CLICKHOUSE_HOSTPORT"),
		Database:   GetEnvOrDefault("PRISME_CLICKHOUSE_DB", "prisme"),
		User:       secret.New(MustGetEnv("PRISME_CLICKHOUSE_USER")),
		Password:   secret.New(MustGetEnv("PRISME_CLICKHOUSE_PASSWORD")),
	}
}
