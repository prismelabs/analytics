package main

import (
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
)

// ProvideConfig is a wire provider config.Config.
func ProvideConfig(bootstrapLogger BootstrapLogger) config.Config {
	logger := log.Logger(bootstrapLogger)

	logger.Info().Msg("loading configuration...")
	cfg := config.FromEnv()
	logger.Info().Any("config", cfg).Msg("configuration successfully loaded.")

	return cfg
}
