package main

import (
	"os"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/rs/zerolog"
)

// ProvideLogger is a wire provider for StandardLogger.
func ProvideLogger() zerolog.Logger {
	logger := log.NewLogger("app", os.Stderr, true)
	log.TestLoggers(logger)

	return logger
}
