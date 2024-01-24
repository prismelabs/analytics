package main

import (
	"fmt"
	"os"

	"github.com/prismelabs/prismeanalytics/cmd/server/full"
	"github.com/prismelabs/prismeanalytics/cmd/server/ingestion"
	"github.com/prismelabs/prismeanalytics/cmd/server/wired"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
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
