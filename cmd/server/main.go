package main

import (
	"fmt"
	"os"

	"github.com/prismelabs/prismeanalytics/internal/log"
)

func main() {
	// Bootstrap logger.
	logger := log.NewLogger("bootstrap", os.Stderr, true)
	log.TestLoggers(logger)

	app := initialize(BootstrapLogger(logger))

	socket := "0.0.0.0:" + fmt.Sprint(app.cfg.Server.Port)
	logger.Info().Msgf("start listening for incoming requests on http://%v", socket)
	logger.Panic().Err(app.fiber.Listen(socket)).Send()
}
