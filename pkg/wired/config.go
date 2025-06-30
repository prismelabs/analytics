package wired

import (
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/rs/zerolog"
)

// ProvideServerConfig is a wire provider for config.Server.
func ProvideServerConfig(bootstrapLogger BootstrapLogger) config.Server {
	logger := zerolog.Logger(bootstrapLogger)

	logger.Info().Msg("loading server configuration...")
	cfg := config.ServerFromEnv()
	logger.Info().Any("config", cfg).Msg("server configuration successfully loaded.")

	return cfg
}
