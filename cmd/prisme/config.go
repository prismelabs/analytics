package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/negrel/configue"
	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/grafana"
	"github.com/prismelabs/analytics/pkg/prisme"
	"github.com/prismelabs/analytics/pkg/services/eventdb"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/originregistry"
	"github.com/prismelabs/analytics/pkg/services/sessionstore"
)

type Config struct {
	prisme         prisme.Config
	chdb           chdb.Config
	clickhouse     clickhouse.Config
	grafana        grafana.Config
	sessionstore   sessionstore.Config
	fiber          fiber.Config
	eventDb        eventdb.Config
	eventStore     eventstore.Config
	originRegistry originregistry.Config
}

func (c *Config) RegisterOptions(figue *configue.Figue) {
	c.prisme.RegisterOptions(figue)
	c.chdb.RegisterOptions(figue)
	c.clickhouse.RegisterOptions(figue)
	c.grafana.RegisterOptions(figue)
	c.sessionstore.RegisterOptions(figue)
	c.eventDb.RegisterOptions(figue)
	c.eventStore.RegisterOptions(figue)
	c.originRegistry.RegisterOptions(figue)
}

func defaultConfig() {
	ini := configue.NewINI(configue.File("./", "config.ini"))
	figue := configue.New("default-config", configue.ContinueOnError, ini)
	var cfg Config
	cfg.RegisterOptions(figue)

	ini.SetOutput(os.Stdout)
	ini.PropSet.PrintDefaults()
}
