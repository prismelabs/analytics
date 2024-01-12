package main

import (
	"os"

	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
)

type BootstrapLogger log.Logger

// ProvideLogger is a wire provider for StandardLogger.
func ProvideLogger(cfg config.Config) log.Logger {
	logger := log.NewLogger("app", os.Stderr, cfg.Server.Debug)
	log.TestLoggers(logger)

	return logger
}
