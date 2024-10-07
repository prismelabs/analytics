package wired

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

// App holds data used at runtime by main package.
type App struct {
	Config          config.Server
	Fiber           *fiber.App
	Logger          zerolog.Logger
	PromLogger      promhttp.Logger
	PromRegistry    *prometheus.Registry
	TeardownService teardown.Service
	setup           Setup
}

// ProvideApp is a wire provider for App.
func ProvideApp(
	app *fiber.App,
	cfg config.Server,
	logger zerolog.Logger,
	promLogger promhttp.Logger,
	promRegistry *prometheus.Registry,
	setup Setup,
	teardownService teardown.Service,
) App {
	return App{
		Config:          cfg,
		Fiber:           app,
		Logger:          logger,
		PromLogger:      promLogger,
		PromRegistry:    promRegistry,
		TeardownService: teardownService,
		setup:           setup,
	}
}
