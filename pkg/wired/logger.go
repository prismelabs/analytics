package wired

import (
	"os"

	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/log"
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
