package main

import (
	"errors"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/negrel/configue"
	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/options"
	"github.com/prismelabs/analytics/pkg/services/eventdb"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/originregistry"
	"github.com/prismelabs/analytics/pkg/services/sessionstore"
)

type Config struct {
	Proxy          options.Proxy
	Server         options.Server
	Admin          options.Admin
	ChDb           chdb.Config
	Clickhouse     clickhouse.Config
	Sessionstore   sessionstore.Config
	Fiber          fiber.Config
	EventDb        eventdb.Config
	EventStore     eventstore.Config
	OriginRegistry originregistry.Config
}

// RegisterOptions registers options in provided Figue.
func (c *Config) RegisterOptions(figue *configue.Figue) {
	c.Proxy.RegisterOptions(figue)
	c.Server.RegisterOptions(figue)
	c.Admin.RegisterOptions(figue)
	c.ChDb.RegisterOptions(figue)
	c.Clickhouse.RegisterOptions(figue)
	c.Sessionstore.RegisterOptions(figue)
	c.EventDb.RegisterOptions(figue)
	c.EventStore.RegisterOptions(figue)
	c.OriginRegistry.RegisterOptions(figue)
}

// Validate validates configuration options.
func (c *Config) Validate() error {
	var errs []error

	// Validated later.
	// c.ChDb.Validate(),
	// c.Clickhouse.Validate(),

	errs = append(errs,
		c.Server.Validate(),
		c.Admin.Validate(),
		// c.Proxy.Validate(),
		c.Sessionstore.Validate(),
		c.EventDb.Validate(),
		c.EventStore.Validate(),
		c.OriginRegistry.Validate())

	switch c.EventDb.Driver {
	case "clickhouse":
		errs = append(errs, c.Clickhouse.Validate())
	case "chdb":
		errs = append(errs, c.ChDb.Validate())
	}

	return errors.Join(errs...)
}

func defaultConfig() {
	ini := configue.NewINI(configFilePath())
	figue := configue.New("default-config", configue.ContinueOnError, ini)
	var cfg Config
	cfg.RegisterOptions(figue)

	ini.SetOutput(os.Stdout)
	ini.PropSet.PrintDefaults()
}

func configFilePath() string {
	if fpath := os.Getenv("PRISME_CONFIG"); fpath != "" {
		return fpath
	}
	return configue.File("./", "config.ini")
}
