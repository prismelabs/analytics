package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/negrel/configue"
	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/prisme"
	"github.com/prismelabs/analytics/pkg/services/eventdb"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/originregistry"
	"github.com/prismelabs/analytics/pkg/services/sessionstore"
)

type Config struct {
	Prisme         prisme.Config
	ChDb           chdb.Config
	Clickhouse     clickhouse.Config
	Sessionstore   sessionstore.Config
	Fiber          fiber.Config
	EventDb        eventdb.Config
	EventStore     eventstore.Config
	OriginRegistry originregistry.Config
}

func (c *Config) RegisterOptions(figue *configue.Figue) {
	c.Prisme.RegisterOptions(figue)
	c.ChDb.RegisterOptions(figue)
	c.Clickhouse.RegisterOptions(figue)
	c.Sessionstore.RegisterOptions(figue)
	c.EventDb.RegisterOptions(figue)
	c.EventStore.RegisterOptions(figue)
	c.OriginRegistry.RegisterOptions(figue)
}

func defaultConfig() {
	ini := configue.NewINI(configue.File("./", "config.ini"))
	figue := configue.New("default-config", configue.ContinueOnError, ini)
	var cfg Config
	cfg.RegisterOptions(figue)

	ini.SetOutput(os.Stdout)
	ini.PropSet.PrintDefaults()
}
