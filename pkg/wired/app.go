package wired

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/rs/zerolog"
)

type App struct {
	Config          config.Server
	Fiber           *fiber.App
	Logger          zerolog.Logger
	TeardownService teardown.Service
	setup           Setup
}

// ProvideApp is a wire provider for App.
func ProvideApp(
	cfg config.Server,
	app *fiber.App,
	logger zerolog.Logger,
	teardownService teardown.Service,
	setup Setup,
) App {
	return App{
		Config:          cfg,
		Fiber:           app,
		Logger:          logger,
		TeardownService: teardownService,
		setup:           setup,
	}
}
