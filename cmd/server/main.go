package main

import (
	"fmt"
	"os"

	"github.com/prismelabs/analytics/cmd/server/full"
	"github.com/prismelabs/analytics/cmd/server/ingestion"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/wired"
)

func main() {
	// Bootstrap logger.
	logger := log.NewLogger("bootstrap", os.Stderr, true)
	log.TestLoggers(logger)

	var app wired.App

	// Initialize server depending on mode.
	mode := config.GetEnvOrDefault("PRISME_MODE", "default")
	switch mode {
	case "ingestion":
		logger.Info().Msg("initilializing ingestion server...")
		app = ingestion.Initialize(wired.BootstrapLogger(logger))
		logger.Info().Msg("ingestion server successfully initialized.")

	case "default":
		logger.Info().Msg("initilializing default server...")
		app = full.Initialize(wired.BootstrapLogger(logger))
		logger.Info().Msg("default server successfully initialized.")

	default:
		logger.Panic().Str("mode", mode).Msg("unknown server mode")
	}

	socket := "0.0.0.0:" + fmt.Sprint(app.Config.Port)
	logger.Info().Msgf("start listening for incoming requests on http://%v", socket)
	logger.Panic().Err(app.Fiber.Listen(socket)).Send()
}
