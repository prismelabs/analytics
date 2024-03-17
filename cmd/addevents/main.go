package main

import (
	"os"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/wired"
)

func main() {
	// Bootstrap logger.
	logger := log.NewLogger("bootstrap", os.Stderr, true)
	log.TestLoggers(logger)

	app := Initialize(wired.BootstrapLogger(logger))
	app.logger.Info().Any("config", app.cfg).Msg("initialization done.")

	if app.cfg.EventType == "pageview" {
		app.AddPageviewsEvents()
	} else {
		app.AddCustomEvents()
	}
}
