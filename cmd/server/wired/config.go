package wired

import (
	"github.com/prismelabs/prismeanalytics/pkg/config"
	"github.com/prismelabs/prismeanalytics/pkg/log"
)

// ProvideClickhouseConfig is a wire provider for config.Clickhouse.
func ProvideClickhouseConfig(bootstrapLogger BootstrapLogger) config.Clickhouse {
	logger := log.Logger(bootstrapLogger)

	logger.Info().Msg("loading clickhouse configuration...")
	cfg := config.ClickhouseFromEnv()
	logger.Info().Any("config", cfg).Msg("clickhouse configuration successfully loaded.")

	return cfg
}

// ProvideServerConfig is a wire provider for config.Server.
func ProvideServerConfig(bootstrapLogger BootstrapLogger) config.Server {
	logger := log.Logger(bootstrapLogger)

	logger.Info().Msg("loading server configuration...")
	cfg := config.ServerFromEnv()
	logger.Info().Any("config", cfg).Msg("server configuration successfully loaded.")

	return cfg
}

// ProvideGrafanaConfig is a wire provider for config.Server.
func ProvideGrafanaConfig(bootstrapLogger BootstrapLogger) config.Grafana {
	logger := log.Logger(bootstrapLogger)

	logger.Info().Msg("loading grafana configuration...")
	cfg := config.GrafanaFromEnv()
	logger.Info().Any("config", cfg).Msg("grafana configuration successfully loaded.")

	return cfg
}
