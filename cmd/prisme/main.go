package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/memory"
	"github.com/negrel/configue"
	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/grafana"
	"github.com/prismelabs/analytics/pkg/handlers"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/middlewares"
	"github.com/prismelabs/analytics/pkg/prisme"
	"github.com/prismelabs/analytics/pkg/services/eventdb"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	grafanaService "github.com/prismelabs/analytics/pkg/services/grafana"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/originregistry"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstore"
	"github.com/prismelabs/analytics/pkg/services/stats"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Setup configuration loader.
	env := configue.NewEnv("PRISME")
	figue := configue.New("", configue.ContinueOnError, env, configue.NewFlag())
	var (
		prismeCfg         prisme.Config
		chdbCfg           chdb.Config
		clickhouseCfg     clickhouse.Config
		grafanaCfg        grafana.Config
		sessionstoreCfg   sessionstore.Config
		fiberCfg          fiber.Config
		eventDbCfg        eventdb.Config
		eventStoreCfg     eventstore.Config
		originRegistryCfg originregistry.Config
	)
	prismeCfg.RegisterOptions(figue)
	chdbCfg.RegisterOptions(figue)
	clickhouseCfg.RegisterOptions(figue)
	grafanaCfg.RegisterOptions(figue)
	sessionstoreCfg.RegisterOptions(figue)
	eventDbCfg.RegisterOptions(figue)
	eventStoreCfg.RegisterOptions(figue)
	originRegistryCfg.RegisterOptions(figue)
	mode := figue.String("mode", "default", "prisme server `mode` (default, ingestion)")

	// Load configuration.
	err := figue.Parse()
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr)
		// `flag` package print usage on flag.ErrHelp so we just prints environment
		// variable defaults.
		env.PrintDefaults()
		return
	}
	if err != nil {
		cliError(figue, err)
	}

	// Validate configuration.
	err = errors.Join(
		prismeCfg.Validate(),
		// Validated later.
		// chdbCfg.Validate(),
		// clickhouseCfg.Validate(),
		// grafanaCfg.Validate(),
		sessionstoreCfg.Validate(),
		eventDbCfg.Validate(),
		eventStoreCfg.Validate(),
		originRegistryCfg.Validate(),
	)
	if err != nil {
		cliError(figue, err)
	}

	// Create application logger.
	logger := log.New("app", os.Stderr, prismeCfg.Debug)
	err = logger.TestOutput()
	if err != nil {
		panic(err)
	}

	if *mode == "default" {
		err := grafanaCfg.Validate()
		if err != nil {
			cliError(figue, err)
		}

		// Setup Grafana dashboard.
		srv := grafanaService.NewService(grafana.NewClient(grafanaCfg), clickhouseCfg)
		err = srv.SetupDatasourceAndDashboards(context.Background(), grafana.OrgId(grafanaCfg.OrgId))
		if err != nil {
			logger.Fatal("failed to setup ClickHouse datasource and dashboard in Grafana", err)
		}
	}

	// Sets event store backend config.
	var driverCfg any
	if eventDbCfg.Driver == "clickhouse" {
		err = clickhouseCfg.Validate()
		driverCfg = clickhouseCfg
	} else {
		err = chdbCfg.Validate()
		driverCfg = chdbCfg
	}
	if err != nil {
		cliError(figue, err)
	}

	// Fiber configuration.
	fiberCfg = fiber.Config{
		ServerHeader:          "prisme",
		StrictRouting:         true,
		AppName:               "Prisme Analytics",
		DisableStartupMessage: true,
		ErrorHandler: func(_ *fiber.Ctx, _ error) error {
			// Errors are handled by errorHandlerMiddleware so access log
			// contains right status code.
			return nil
		},
	}
	if prismeCfg.TrustProxy {
		fiberCfg.EnableIPValidation = false
		fiberCfg.ProxyHeader = prismeCfg.ProxyHeader
	} else {
		fiberCfg.EnableIPValidation = true
		fiberCfg.ProxyHeader = ""
	}

	// Setup prometheus registry.
	promRegistry := prometheus.NewRegistry()
	// Collectors of default prometheus registry.
	promRegistry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	promRegistry.MustRegister(collectors.NewGoCollector())

	// Create teardown service.
	teardownService := teardown.NewService()

	// Setup services.
	eventDb, err := eventdb.NewService(
		eventDbCfg,
		driverCfg,
		logger,
		clickhouse.EmbeddedSourceDriver(logger),
		teardownService,
	)
	if err != nil {
		cliError(figue, err)
	}
	eventStore, err := eventstore.NewService(
		eventStoreCfg,
		eventDb,
		logger,
		promRegistry,
		teardownService,
	)
	if err != nil {
		cliError(figue, err)
	}
	stats := stats.NewService(eventDb)
	uaParser := uaparser.NewService(logger, promRegistry)
	ipGeolocator := ipgeolocator.NewMmdbService(logger, promRegistry)
	saltManager := saltmanager.NewService(logger)
	sessionStore := sessionstore.NewService(logger, sessionstoreCfg, promRegistry)
	originRegistry := originregistry.NewService(originRegistryCfg, logger)

	// Create fiber app.
	app := fiber.New(fiberCfg)

	teardownService.RegisterProcedure(func() error {
		logger.Info("shutting down fiber server...")
		err := app.Shutdown()
		if err != nil {
			logger.Err("failed to shutdown fiber server", err)
		} else {
			logger.Info("fiber server shutdown")
		}

		return err
	})

	app.Use(middlewares.Metrics(promRegistry),
		middlewares.RequestId(prismeCfg),
		middlewares.AccessLog(prismeCfg, logger),
		middlewares.ErrorHandler(promRegistry, logger))

	// Register handlers.
	{
		// Public endpoints.
		app.Use("/static", handlers.Static(prismeCfg))

		app.Use("/api/v1/healthcheck", handlers.HealthCheck())

		eventCors := middlewares.EventsCors()
		eventRateLimit := middlewares.EventsRateLimiter(
			prismeCfg,
			memory.New(memory.Config{
				GCInterval: 10 * time.Second,
			}),
		)
		nonRegisteredOriginFilter := middlewares.NonRegisteredOriginFilter(originRegistry)
		eventTimeout := middlewares.ApiEventsTimeout(prismeCfg)

		app.Use("/api/v1/events/*",
			eventCors,
			eventRateLimit,
			nonRegisteredOriginFilter,
			eventTimeout,
		)

		app.Use("/api/v1/noscript/events/*",
			eventCors,
			eventRateLimit,
			nonRegisteredOriginFilter,
			eventTimeout,
			// Prevent caching of GET responses.
			middlewares.NoscriptHandlersCache(),
		)

		app.Post("/api/v1/events/pageviews",
			handlers.PostEventsPageViews(
				eventStore,
				uaParser,
				ipGeolocator,
				saltManager,
				sessionStore,
			),
		)
		app.Get("/api/v1/noscript/events/pageviews",
			handlers.GetNoscriptEventsPageviews(
				eventStore,
				uaParser,
				ipGeolocator,
				saltManager,
				sessionStore,
			),
		)

		app.Post("/api/v1/events/custom/:name",
			fiber.Handler(handlers.PostEventsCustom(
				eventStore,
				saltManager,
				sessionStore,
			)),
		)
		app.Get("/api/v1/noscript/events/custom/:name",
			handlers.GetNoscriptEventsCustom(eventStore,
				saltManager,
				sessionStore,
			),
		)

		app.Post("/api/v1/events/outbound-links",
			handlers.PostEventsOutboundLinks(
				eventStore,
				saltManager,
				sessionStore,
			),
		)
		app.Get("/api/v1/noscript/events/outbound-links",
			handlers.GetNoscriptEventsOutboundLinks(
				eventStore,
				sessionStore,
				saltManager,
			),
		)

		app.Post("/api/v1/events/file-downloads",
			handlers.PostEventsFileDownloads(
				eventStore,
				saltManager,
				sessionStore,
			),
		)

		app.Use("/api/v1/stats/*", middlewares.StatsCors(prismeCfg))
		app.Get("/api/v1/stats/batch", handlers.GetStatsBatch(stats, logger))
	}

	// Admin and profiling server.
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := io.WriteString(w, "consult metrics at /metrics")
			if err != nil {
				logger.Err("failed to write admin response body", err)
			}
		})
		http.Handle("/metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{
			ErrorLog:            log.PrometheusLogger(logger),
			ErrorHandling:       promhttp.HTTPErrorOnError,
			Registry:            promRegistry,
			DisableCompression:  false,
			MaxRequestsInFlight: 0,
			Timeout:             3 * time.Second,
			EnableOpenMetrics:   false,
			ProcessStartTime:    time.Now(),
		}))
		logger.Info("admin server listening for incoming request", "host_port", prismeCfg.AdminHostPort)
		err := http.ListenAndServe(prismeCfg.AdminHostPort, nil)
		logger.Fatal("failed to start admin server", err)
	}()

	go func() {
		socket := "0.0.0.0:" + fmt.Sprint(prismeCfg.Port)
		logger.Info("start listening for incoming requests", "host_port", socket)
		err := app.Listen(socket)
		logger.Fatal("failed to listen for incoming requests", err)
	}()

	ch := make(chan os.Signal, 16)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch

	logger.Info("starting tearing down procedures...")
	err = teardownService.Teardown()
	logger.Fatal("tearing down procedures done.", err)
	logger.Info("tearing down successful, exiting...")
}

func cliError(figue *configue.Figue, err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	figue.PrintDefaults()
	os.Exit(1)
}
