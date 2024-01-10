package main

import (
	"github.com/labstack/echo/v4"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
)

// ProvideEcho is a wire provider for echo.Echo. It setup middleware and handlers.
func ProvideEcho(cfg config.Config, accessLogger AccessLogger) *echo.Echo {
	e := echo.New()
	if cfg.Server.TrustProxy {
		e.IPExtractor = echo.ExtractIPFromXFFHeader()
	} else {
		e.IPExtractor = echo.ExtractIPDirect()
	}
	e.HideBanner = true
	e.HidePort = true

	e.Use(middlewares.RequestId(cfg.Server))
	e.Use(middlewares.AccessLog(accessLogger.Logger))

	return e
}
