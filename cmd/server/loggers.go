package main

import (
	"os"

	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
)

// StandardLogger is the application logger.
type StandardLogger struct {
	log.Logger
}

// ProvideStandardLogger is a wire provider for StandardLogger.
func ProvideStandardLogger(cfg config.Config) StandardLogger {
	logger := log.NewLogger("app", os.Stderr, cfg.Server.Debug)
	log.TestLoggers(logger)

	return StandardLogger{logger}
}

type AccessLogger struct {
	log.Logger
}

// ProvideAccessLogger is a wire provider for AccessLogger.
func ProvideAccessLogger(cfg config.Config, logger StandardLogger) AccessLogger {
	// Open access log file.
	accessLogFile, err := os.OpenFile(cfg.Server.AccessLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		logger.Panic().Err(err).Msgf("failed to open access log file: %v", cfg.Server.AccessLog)
	}

	accessLogger := log.NewLogger("access_log", accessLogFile, cfg.Server.Debug)
	log.TestLoggers(accessLogger)

	return AccessLogger{accessLogger}
}
