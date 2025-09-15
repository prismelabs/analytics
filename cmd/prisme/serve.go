package main

import (
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
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/handlers"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/middlewares"
	"github.com/prismelabs/analytics/pkg/services/eventdb"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
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

func serve() {
	// Setup configuration loader.
	ini := configue.NewINI(configue.File("./", "config.ini"))
	figue := configue.New(
		"",
		configue.ContinueOnError,
		ini,
		configue.NewEnv("PRISME"),
		configue.NewFlag(),
	)
	figue.Usage = func() {
		_, _ = fmt.Fprintln(figue.Output(), "prisme - High-perfomance, self-hosted and privacy-focused web analytics service.")
		_, _ = fmt.Fprintln(figue.Output())
		_, _ = fmt.Fprintln(figue.Output(), "Usage:")
		_, _ = fmt.Fprintln(figue.Output(), "  prisme [COMMAND] [FLAGS]")
		_, _ = fmt.Fprintln(figue.Output())
		_, _ = fmt.Fprintln(figue.Output(), "  prisme serve -eventdb-driver chdb -chdb-path ./prisme -origins 'localhost,prismeanalytics.com'")
		_, _ = fmt.Fprintln(figue.Output())
		_, _ = fmt.Fprintln(figue.Output(), "Commands:")
		_, _ = fmt.Fprintln(figue.Output(), "  serve")
		_, _ = fmt.Fprintln(figue.Output(), "        start web analytics server, this is the default")
		_, _ = fmt.Fprintln(figue.Output(), "  grafana-dashboard")
		_, _ = fmt.Fprintln(figue.Output(), "        generate and print grafana dashboard to stdout")
		_, _ = fmt.Fprintln(figue.Output(), "  default-config")
		_, _ = fmt.Fprintln(figue.Output(), "        print default configuration file to stdout")
		_, _ = fmt.Fprintln(figue.Output())
		figue.PrintDefaults()
	}

	var cfg Config
	cfg.RegisterOptions(figue)

	// Load configuration.
	err := figue.Parse()
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		cliError(err)
	}

	// Validate configuration.
	err = errors.Join(
		cfg.prisme.Validate(),
		// Validated later.
		// cfg.chdb.Validate(),
		// cfg.clickhouse.Validate(),
		// cfg.grafana.Validate(),
		cfg.sessionstore.Validate(),
		cfg.eventDb.Validate(),
		cfg.eventStore.Validate(),
		cfg.originRegistry.Validate(),
	)
	if err != nil {
		cliError(err)
	}

	// Create application logger.
	logger := log.New("app", os.Stderr, cfg.prisme.Debug)
	err = logger.TestOutput()
	if err != nil {
		panic(err)
	}

	// Sets event store backend config.
	var driverCfg any
	if cfg.eventDb.Driver == "clickhouse" {
		err = cfg.clickhouse.Validate()
		driverCfg = cfg.clickhouse
	} else {
		err = cfg.chdb.Validate()
		driverCfg = cfg.chdb
	}
	if err != nil {
		cliError(err)
	}

	// Fiber configuration.
	cfg.fiber = fiber.Config{
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
	if cfg.prisme.TrustProxy {
		cfg.fiber.EnableIPValidation = false
		cfg.fiber.ProxyHeader = cfg.prisme.ProxyHeader
	} else {
		cfg.fiber.EnableIPValidation = true
		cfg.fiber.ProxyHeader = ""
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
		cfg.eventDb,
		driverCfg,
		logger,
		clickhouse.EmbeddedSourceDriver(logger),
		teardownService,
	)
	if err != nil {
		cliError(err)
	}
	eventStore, err := eventstore.NewService(
		cfg.eventStore,
		eventDb,
		logger,
		promRegistry,
		teardownService,
	)
	if err != nil {
		cliError(err)
	}
	stats := stats.NewService(eventDb, teardownService)
	uaParser := uaparser.NewService(logger, promRegistry)
	ipGeolocator := ipgeolocator.NewMmdbService(logger, promRegistry)
	saltManager := saltmanager.NewService(logger)
	sessionStore := sessionstore.NewService(logger, cfg.sessionstore, promRegistry)
	originRegistry, err := originregistry.NewService(cfg.originRegistry, logger)
	if err != nil {
		cliError(err)
	}

	// Create fiber app.
	app := fiber.New(cfg.fiber)

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
		middlewares.RequestId(cfg.prisme),
		middlewares.AccessLog(cfg.prisme, logger),
		middlewares.ErrorHandler(promRegistry, logger))

	// Register handlers.
	{
		// Public endpoints.

		app.Use("/static", handlers.Static(cfg.prisme))

		app.Use("/dashboard", handlers.Dashboard())

		app.Use("/api/v1/healthcheck", handlers.HealthCheck())

		eventCors := middlewares.EventsCors()
		eventRateLimit := middlewares.EventsRateLimiter(
			cfg.prisme,
			memory.New(memory.Config{
				GCInterval: 10 * time.Second,
			}),
		)
		nonRegisteredOriginFilter := middlewares.NonRegisteredOriginFilter(originRegistry)
		eventTimeout := middlewares.ApiEventsTimeout(cfg.prisme)

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

		stats := handlers.GetStatsHandlers(stats)
		app.Use("/api/v1/stats/*", middlewares.StatsCors(cfg.prisme))
		app.Get("/api/v1/stats/bounces", stats.Bounces)
		app.Get("/api/v1/stats/visitors", stats.Visitors)
		app.Get("/api/v1/stats/sessions", stats.Sessions)
		app.Get("/api/v1/stats/sessions-duration", stats.SessionsDuration)
		app.Get("/api/v1/stats/pageviews", stats.PageViews)
		app.Get("/api/v1/stats/live-visitors", stats.LiveVisitors)
		app.Get("/api/v1/stats/top-pages", stats.TopPages)
		app.Get("/api/v1/stats/top-entry-pages", stats.TopEntryPages)
		app.Get("/api/v1/stats/top-exit-pages", stats.TopExitPages)
		app.Get("/api/v1/stats/top-referrers", stats.TopReferrers)
		app.Get("/api/v1/stats/top-utm-sources", stats.TopUtmSources)
		app.Get("/api/v1/stats/top-utm-mediums", stats.TopUtmMediums)
		app.Get("/api/v1/stats/top-utm-campaigns", stats.TopUtmCampaigns)
		app.Get("/api/v1/stats/top-countries", stats.TopCountries)
		app.Get("/api/v1/stats/top-operating-systems", stats.TopOperatingSystems)
		app.Get("/api/v1/stats/top-browsers", stats.TopBrowsers)
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
		logger.Info("admin server listening for incoming request", "host_port", cfg.prisme.AdminHostPort)
		err := http.ListenAndServe(cfg.prisme.AdminHostPort, nil)
		logger.Fatal("failed to start admin server", err)
	}()

	go func() {
		socket := "0.0.0.0:" + fmt.Sprint(cfg.prisme.Port)
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
