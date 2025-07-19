package main

import (
	"os"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/rs/zerolog"
)

// NewLogger returns a new configured logger.
func NewLogger() zerolog.Logger {
	logger := log.NewLogger("app", os.Stderr, true)
	log.TestLoggers(logger)

	return logger
}
