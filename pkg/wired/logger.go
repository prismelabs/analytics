package wired

import (
	"os"

	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/log"
)

type BootstrapLogger log.Logger

// ProvideLogger is a wire provider for StandardLogger.
func ProvideLogger(cfg config.Server) log.Logger {
	logger := log.NewLogger("app", os.Stderr, cfg.Debug)
	log.TestLoggers(logger)

	return logger
}
