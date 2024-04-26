package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prismelabs/analytics/cmd/server/full"
	"github.com/prismelabs/analytics/cmd/server/ingestion"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/wired"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Bootstrap logger.
	logger := log.NewLogger("bootstrap", os.Stderr, true)
	log.TestLoggers(logger)

	var app wired.App

	// Initialize server depending on mode.
	mode := config.GetEnvOrDefault("PRISME_MODE", "default")
	switch mode {
	case "ingestion":
		logger.Info().Msg("initilializing ingestion server...")
		app = ingestion.Initialize(wired.BootstrapLogger(logger))
		logger.Info().Msg("ingestion server successfully initialized.")

	case "default":
		logger.Info().Msg("initilializing default server...")
		app = full.Initialize(wired.BootstrapLogger(logger))
		app.Logger.Info().Msg("default server successfully initialized.")

	default:
		app.Logger.Panic().Str("mode", mode).Msg("unknown server mode")
	}

	// Admin and profiling server.
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(app.PromRegistry, promhttp.HandlerOpts{
			ErrorLog:            &app.Logger,
			ErrorHandling:       promhttp.HTTPErrorOnError,
			Registry:            app.PromRegistry,
			DisableCompression:  false,
			MaxRequestsInFlight: 0,
			Timeout:             3 * time.Second,
			EnableOpenMetrics:   false,
			ProcessStartTime:    time.Now(),
		}))
		app.Logger.Info().Msgf("admin server listening for incoming request on http://%v", app.Config.AdminHostPort)
		err := http.ListenAndServe(app.Config.AdminHostPort, nil)
		app.Logger.Panic().Err(err).Msg("failed to start admin server")
	}()

	go func() {
		socket := "0.0.0.0:" + fmt.Sprint(app.Config.Port)
		app.Logger.Info().Msgf("start listening for incoming requests on http://%v", socket)
		err := app.Fiber.Listen(socket)
		if err != nil {
			app.Logger.Panic().Err(err).Send()
		}
	}()

	ch := make(chan os.Signal, 16)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch

	app.Logger.Info().Msg("starting tearing down procedures...")
	err := app.TeardownService.Teardown()
	app.Logger.Err(err).Msg("tearing down procedures done.")
}
