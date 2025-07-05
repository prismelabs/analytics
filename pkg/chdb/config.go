package chdb

import (
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/rs/zerolog"
)

// chdb options.
type Config struct {
	Path string
}

// configFromEnv loads chdb config from environment variables.
func configFromEnv() Config {
	return Config{
		Path: config.MustGetEnv("PRISME_CHDB_PATH"),
	}
}

// ProvideConfig is a wire provider for config.
func ProvideConfig(logger zerolog.Logger) Config {
	logger.Info().Msg("loading chdb configuration...")
	cfg := configFromEnv()
	logger.Info().Any("config", cfg).Msg("chdb configuration successfully loaded.")

	return cfg
}
