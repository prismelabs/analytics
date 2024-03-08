package wired

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/rs/zerolog"
)

type App struct {
	Config config.Server
	Fiber  *fiber.App
	Logger zerolog.Logger
	setup  Setup
}

// ProvideApp is a wire provider for App.
func ProvideApp(cfg config.Server, app *fiber.App, logger zerolog.Logger, setup Setup) App {
	return App{
		Config: cfg,
		Fiber:  app,
		Logger: logger,
		setup:  setup,
	}
}
