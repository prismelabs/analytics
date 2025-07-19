package main

import (
	"os"
	"time"

	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/teardown"
)

func main() {
	// Bootstrap logger.
	logger := log.NewLogger("bootstrap", os.Stderr, true)
	log.TestLoggers(logger)

	zerologLogger := ProvideLogger()
	config := ProvideConfig()
	driver := clickhouse.ProvideEmbeddedSourceDriver(zerologLogger)
	service := teardown.ProvideService()
	app := ProvideApp(zerologLogger, config, driver, service)

	app.logger.Info().Any("config", app.cfg).Msg("initialization done.")

	start := time.Now()

	app.executeScenario(emulateSession)

	app.logger.Info().
		Stringer("duration", time.Since(start)).Msg("scenario done")
}
