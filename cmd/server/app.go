package main

import (
	"github.com/labstack/echo/v4"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
)

type App struct {
	cfg    config.Config
	echo   *echo.Echo
	logger log.Logger
}

// ProvideApp is a wire provider for App.
func ProvideApp(cfg config.Config, e *echo.Echo, logger StandardLogger) App {
	return App{
		cfg:    cfg,
		echo:   e,
		logger: logger.Logger,
	}
}
