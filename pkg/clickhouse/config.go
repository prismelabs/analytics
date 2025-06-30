package clickhouse

import (
	"github.com/negrel/secrecy"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/rs/zerolog"
)

// Clickhouse connection options.
type Config struct {
	TlsEnabled bool
	HostPort   string
	Database   string
	User       secrecy.SecretString
	Password   secrecy.SecretString
}

// configFromEnv loads clickhouse config from environment variables.
// This function panics if required environment variables are missing.
func configFromEnv() Config {
	return Config{
		TlsEnabled: config.GetEnvOrDefault("PRISME_CLICKHOUSE_TLS", "false") != "false",
		HostPort:   config.MustGetEnv("PRISME_CLICKHOUSE_HOSTPORT"),
		Database:   config.GetEnvOrDefault("PRISME_CLICKHOUSE_DB", "prisme"),
		User:       secrecy.NewSecretString(secrecy.UnsafeStringToBytes(config.MustGetEnv("PRISME_CLICKHOUSE_USER"))),
		Password:   secrecy.NewSecretString(secrecy.UnsafeStringToBytes(config.MustGetEnv("PRISME_CLICKHOUSE_PASSWORD"))),
	}
}

// ProvideConfig is a wire provider for config.
func ProvideConfig(logger zerolog.Logger) Config {
	logger.Info().Msg("loading clickhouse configuration...")
	cfg := configFromEnv()
	logger.Info().Any("config", cfg).Msg("clickhouse configuration successfully loaded.")

	return cfg
}
