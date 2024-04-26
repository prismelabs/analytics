package wired

import (
	"os"

	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// BootstrapLogger define zerolog loggers used before ProvideLogger.
type BootstrapLogger zerolog.Logger

// ProvideLogger is a wire provider for zerolog.Logger.
func ProvideLogger(cfg config.Server) zerolog.Logger {
	logger := log.NewLogger("app", os.Stderr, cfg.Debug)
	log.TestLoggers(logger)

	return logger
}

// ProvidePromHttpLogger is a wire provider for promhttp.Logger.
func ProvidePromHttpLogger(cfg config.Server, logger zerolog.Logger) promhttp.Logger {
	// Open access log file.
	accessLogFile, err := os.OpenFile(cfg.AccessLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		logger.Panic().Err(err).Msgf("failed to open access log file: %v", cfg.AccessLog)
	}

	accessLogger := log.NewLogger("admin_access_log", accessLogFile, cfg.Debug)
	log.TestLoggers(accessLogger)

	return &accessLogger
}
