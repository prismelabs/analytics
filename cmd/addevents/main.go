package main

import (
	"time"

	"github.com/negrel/configue"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/services/teardown"
)

func main() {
	logger := NewLogger()

	figue := configue.New("", configue.ContinueOnError, configue.NewEnv("PRISME"), configue.NewFlag())
	var (
		config        Config
		clickhouseCfg clickhouse.Config
	)
	config.RegisterOptions(figue)
	clickhouseCfg.RegisterOptions(figue)

	err := figue.Parse()
	if err != nil {
		logger.Fatal("failed to parse configuration options", err)
	}
	err = clickhouseCfg.Validate()
	if err != nil {
		logger.Fatal("invalid options", err)
	}

	driver := clickhouse.EmbeddedSourceDriver(logger)
	teardown := teardown.NewService()
	ch := clickhouse.NewCh(logger, clickhouseCfg, driver, teardown)
	app := NewApp(logger, config, ch)

	app.logger.Info("initialization done", "config", app.cfg)

	start := time.Now()

	app.executeScenario(emulateSession)

	app.logger.Info("scenario done", "duration", time.Since(start))
}
