package main

import (
	"time"

	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/services/teardown"
)

func main() {
	logger := NewLogger()
	config := NewConfig()
	driver := clickhouse.EmbeddedSourceDriver(logger)
	service := teardown.NewService()
	app := NewApp(logger, config, driver, service)

	app.logger.Info("initialization done", "config", app.cfg)

	start := time.Now()

	app.executeScenario(emulateSession)

	app.logger.Info("scenario done", "duration", time.Since(start))
}
