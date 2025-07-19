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
	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/grafana"
	"github.com/prismelabs/analytics/pkg/handlers"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/middlewares"
	"github.com/prismelabs/analytics/pkg/prisme"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/originregistry"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstore"
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
		eventStoreCfg     eventstore.Config
		originRegistryCfg originregistry.Config
	)
	prismeCfg.RegisterOptions(figue)
	chdbCfg.RegisterOptions(figue)
	clickhouseCfg.RegisterOptions(figue)
	grafanaCfg.RegisterOptions(figue)
	sessionstoreCfg.RegisterOptions(figue)
	eventStoreCfg.RegisterOptions(figue)
	originRegistryCfg.RegisterOptions(figue)

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
		fmt.Fprintln(os.Stderr, err.Error())
		figue.PrintDefaults()
		os.Exit(1)
	}

	// Validate configuration.
	err = errors.Join(
		prismeCfg.Validate(),
		// chdbCfg.Validate(),
		// clickhouseCfg.Validate(),
		grafanaCfg.Validate(),
		sessionstoreCfg.Validate(),
		eventStoreCfg.Validate(),
		originRegistryCfg.Validate(),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		figue.PrintDefaults()
		os.Exit(1)
	}

	// Sets eventstore backend config.
	if eventStoreCfg.Backend == "clickhouse" {
		eventStoreCfg.BackendConfig = clickhouseCfg
		err = clickhouseCfg.Validate()
	} else {
		eventStoreCfg.BackendConfig = chdbCfg
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		figue.PrintDefaults()
		os.Exit(1)
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

	// Create application logger.
	logger := log.NewLogger("app", os.Stderr, prismeCfg.Debug)
	log.TestLoggers(logger)

	// Create access logger.
	var accessLogWriter io.Writer
	switch prismeCfg.AccessLog {
	case "/dev/stdout":
		accessLogWriter = os.Stdout
	case "/dev/stderr":
		accessLogWriter = os.Stderr
	default:
		f, err := os.OpenFile(
			prismeCfg.AccessLog,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			os.ModePerm,
		)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to open access log file")
		}
		accessLogWriter = f
	}
	accessLogger := log.NewLogger("access_log", accessLogWriter, false)

	// Setup prometheus registry.
	promRegistry := prometheus.NewRegistry()
	// Collectors of default prometheus registry.
	promRegistry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	promRegistry.MustRegister(collectors.NewGoCollector())

	// Create teardown service.
	teardownService := teardown.ProvideService()

	// Setup some services.
	eventStore := eventstore.ProvideService(
		eventStoreCfg,
		logger,
		promRegistry,
		teardownService,
		clickhouse.ProvideEmbeddedSourceDriver(logger),
	)
	uaParser := uaparser.ProvideService(logger, promRegistry)
	ipGeolocator := ipgeolocator.ProvideMmdbService(logger, promRegistry)
	saltManager := saltmanager.ProvideService(logger)
	sessionStore := sessionstore.ProvideService(logger, sessionstoreCfg, promRegistry)
	originRegistry := originregistry.ProvideEnvVarService(originRegistryCfg, logger)

	// Create fiber app.
	app := fiber.New(fiberCfg)

	teardownService.RegisterProcedure(func() error {
		logger.Info().Msg("shutting down fiber server...")
		err := app.Shutdown()
		logger.Info().Err(err).Msg("fiber server shutdown.")

		return err
	})

	app.Use(fiber.Handler(middlewares.ProvideMetrics(promRegistry)),
		fiber.Handler(middlewares.ProvideRequestId(prismeCfg)),
		fiber.Handler(middlewares.ProvideAccessLog(prismeCfg, accessLogger)),
		fiber.Handler(middlewares.ProvideErrorHandler(promRegistry, logger)))

	// Register handlers.
	{
		// Public endpoints.
		app.Use("/static", fiber.Handler(middlewares.ProvideStatic(prismeCfg)))

		app.Use("/api/v1/healthcheck", fiber.Handler(handlers.ProvideHealthCheck()))

		eventCors := middlewares.ProvideEventsCors()
		eventRateLimit := middlewares.ProvideEventsRateLimiter(
			prismeCfg,
			memory.New(memory.Config{
				GCInterval: 10 * time.Second,
			}),
		)
		nonRegisteredOriginFilter := middlewares.ProvideNonRegisteredOriginFilter(originRegistry)
		eventTimeout := middlewares.ProvideApiEventsTimeout(prismeCfg)

		app.Use("/api/v1/events/*",
			fiber.Handler(eventCors),
			fiber.Handler(eventRateLimit),
			fiber.Handler(nonRegisteredOriginFilter),
			fiber.Handler(eventTimeout),
		)

		app.Use("/api/v1/noscript/events/*",
			fiber.Handler(eventCors),
			fiber.Handler(eventRateLimit),
			fiber.Handler(nonRegisteredOriginFilter),
			fiber.Handler(eventTimeout),
			// Prevent caching of GET responses.
			fiber.Handler(middlewares.ProvideNoscriptHandlersCache()),
		)

		app.Post("/api/v1/events/pageviews",
			fiber.Handler(handlers.ProvidePostEventsPageViews(
				logger,
				eventStore,
				uaParser,
				ipGeolocator,
				saltManager,
				sessionStore,
			)),
		)
		app.Get("/api/v1/noscript/events/pageviews",
			fiber.Handler(handlers.ProvideGetNoscriptEventsPageviews(
				logger,
				eventStore,
				uaParser,
				ipGeolocator,
				saltManager,
				sessionStore,
			)),
		)

		app.Post("/api/v1/events/custom/:name",
			fiber.Handler(handlers.ProvidePostEventsCustom(
				eventStore,
				saltManager,
				sessionStore,
			)),
		)
		app.Get("/api/v1/noscript/events/custom/:name",
			fiber.Handler(handlers.ProvideGetNoscriptEventsCustom(eventStore,
				saltManager,
				sessionStore,
			)),
		)

		app.Post("/api/v1/events/outbound-links",
			fiber.Handler(handlers.ProvidePostEventsOutboundLinks(
				eventStore,
				saltManager,
				sessionStore,
			)),
		)
		app.Get("/api/v1/noscript/events/outbound-links",
			fiber.Handler(handlers.ProvidePostEventsOutboundLinks(
				eventStore,
				saltManager,
				sessionStore,
			)),
		)

		app.Post("/api/v1/events/file-downloads",
			fiber.Handler(handlers.ProvidePostEventsFileDownloads(
				eventStore,
				saltManager,
				sessionStore,
			)),
		)
	}

	// Admin and profiling server.
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := io.WriteString(w, "consult metrics at /metrics")
			if err != nil {
				logger.Err(err).Msg("failed to write admin response body")
			}
		})
		http.Handle("/metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{
			ErrorLog:            &logger,
			ErrorHandling:       promhttp.HTTPErrorOnError,
			Registry:            promRegistry,
			DisableCompression:  false,
			MaxRequestsInFlight: 0,
			Timeout:             3 * time.Second,
			EnableOpenMetrics:   false,
			ProcessStartTime:    time.Now(),
		}))
		logger.Info().Msgf("admin server listening for incoming request on http://%v", prismeCfg.AdminHostPort)
		err := http.ListenAndServe(prismeCfg.AdminHostPort, nil)
		logger.Panic().Err(err).Msg("failed to start admin server")
	}()

	go func() {
		socket := "0.0.0.0:" + fmt.Sprint(prismeCfg.Port)
		logger.Info().Msgf("start listening for incoming requests on http://%v", socket)
		err := app.Listen(socket)
		if err != nil {
			logger.Panic().Err(err).Send()
		}
	}()

	ch := make(chan os.Signal, 16)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch

	logger.Info().Msg("starting tearing down procedures...")
	err = teardownService.Teardown()
	logger.Err(err).Msg("tearing down procedures done.")
}
