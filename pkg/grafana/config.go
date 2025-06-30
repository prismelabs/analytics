package grafana

import (
	"github.com/negrel/secrecy"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/rs/zerolog"
)

// Grafana related options.
type Config struct {
	Url      string
	User     secrecy.SecretString
	Password secrecy.SecretString
	OrgId    int64
}

// configFromEnv loads grafana related options from environment variables.
// This function panics if required environment variables are missing.
func configFromEnv() Config {
	return Config{
		Url:      config.MustGetEnv("PRISME_GRAFANA_URL"),
		User:     secrecy.NewSecretString(secrecy.UnsafeStringToBytes(config.MustGetEnv("PRISME_GRAFANA_USER"))),
		Password: secrecy.NewSecretString(secrecy.UnsafeStringToBytes(config.MustGetEnv("PRISME_GRAFANA_PASSWORD"))),
		OrgId:    config.ParseIntEnvOrDefault("PRISME_GRAFANA_ORG_ID", 1, 64),
	}
}

// ProvideConfig is a wire provider for config.
func ProvideConfig(logger zerolog.Logger) Config {
	logger.Info().Msg("loading grafana configuration...")
	cfg := configFromEnv()
	logger.Info().Any("config", cfg).Msg("grafana configuration successfully loaded.")

	return cfg
}
