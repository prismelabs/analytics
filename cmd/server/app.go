package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
)

type App struct {
	cfg    config.Config
	fiber  *fiber.App
	logger log.Logger
}

// ProvideApp is a wire provider for App.
func ProvideApp(cfg config.Config, app *fiber.App, logger log.Logger) App {
	return App{
		cfg:    cfg,
		fiber:  app,
		logger: logger,
	}
}
