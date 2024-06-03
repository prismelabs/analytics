package main

import (
	"os"
	"time"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/wired"
)

func main() {
	// Bootstrap logger.
	logger := log.NewLogger("bootstrap", os.Stderr, true)
	log.TestLoggers(logger)

	app := Initialize(wired.BootstrapLogger(logger))
	app.logger.Info().Any("config", app.cfg).Msg("initialization done.")

	start := time.Now()

	app.executeScenario(emulateSession)

	app.logger.Info().
		Stringer("duration", time.Since(start)).Msg("scenario done")
}
