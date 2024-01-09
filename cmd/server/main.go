package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
)

func main() {
	// Bootstrap logger.
	logger := log.NewLogger("bootstrap", os.Stderr, true)
	log.TestLoggers(logger)

	logger.Info().Msg("loading configuration...")
	cfg := config.FromEnv()
	logger.Info().Any("config", cfg).Msg("configuration successfully loaded.")

	// Application logger.
	logger = log.NewLogger("app", os.Stderr, cfg.Server.Debug)
	log.TestLoggers(logger)

	// Open access log file.
	accessLogFile, err := os.OpenFile(cfg.Server.AccessLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		logger.Panic().Err(err).Msgf("failed to open access log file: %v", cfg.Server.AccessLog)
	}
	accessLogger := log.NewLogger("access_log", accessLogFile, cfg.Server.Debug)
	log.TestLoggers(logger)

	e := echo.New()
	if cfg.Server.TrustProxy {
		e.IPExtractor = echo.ExtractIPFromXFFHeader()
	} else {
		e.IPExtractor = echo.ExtractIPDirect()
	}
	e.HideBanner = true
	e.HidePort = true

	e.Use(middlewares.RequestId(cfg.Server))
	e.Use(middlewares.AccessLog(accessLogger))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	socket := "0.0.0.0:" + fmt.Sprint(cfg.Server.Port)
	logger.Info().Msgf("start listening for incoming requests on http://%v", socket)
	logger.Panic().Err(e.Start(socket)).Send()
}
