package config

import (
	"github.com/negrel/secrecy"
)

// Clickhouse connection options.
type Clickhouse struct {
	TlsEnabled bool
	HostPort   string
	Database   string
	User       secrecy.SecretString
	Password   secrecy.SecretString
}

// ClickhouseFromEnv loads clickhouse config from environment variables.
// This function panics if required environment variables are missing.
func ClickhouseFromEnv() Clickhouse {
	return Clickhouse{
		TlsEnabled: GetEnvOrDefault("PRISME_CLICKHOUSE_TLS", "false") != "false",
		HostPort:   MustGetEnv("PRISME_CLICKHOUSE_HOSTPORT"),
		Database:   GetEnvOrDefault("PRISME_CLICKHOUSE_DB", "prisme"),
		User:       secrecy.NewSecretString(secrecy.UnsafeStringToBytes(MustGetEnv("PRISME_CLICKHOUSE_USER"))),
		Password:   secrecy.NewSecretString(secrecy.UnsafeStringToBytes(MustGetEnv("PRISME_CLICKHOUSE_PASSWORD"))),
	}
}
