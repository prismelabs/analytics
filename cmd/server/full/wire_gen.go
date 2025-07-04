// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package full

import (
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/grafana"
	"github.com/prismelabs/analytics/pkg/handlers"
	"github.com/prismelabs/analytics/pkg/middlewares"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	grafana2 "github.com/prismelabs/analytics/pkg/services/grafana"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/originregistry"
	"github.com/prismelabs/analytics/pkg/services/saltmanager"
	"github.com/prismelabs/analytics/pkg/services/sessionstore"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prismelabs/analytics/pkg/wired"
)

// Injectors from wire.go:

func Initialize(logger wired.BootstrapLogger) wired.App {
	server := wired.ProvideServerConfig(logger)
	apiEventsTimeout := middlewares.ProvideApiEventsTimeout(server)
	eventsCors := middlewares.ProvideEventsCors()
	storage := wired.ProvideFiberStorage()
	eventsRateLimiter := middlewares.ProvideEventsRateLimiter(server, storage)
	zerologLogger := wired.ProvideLogger(server)
	config := eventstore.ProvideConfig(zerologLogger)
	registry := wired.ProvidePrometheusRegistry()
	service := teardown.ProvideService()
	driver := clickhouse.ProvideEmbeddedSourceDriver(zerologLogger)
	eventstoreService := eventstore.ProvideService(config, zerologLogger, registry, service, driver)
	saltmanagerService := saltmanager.ProvideService(zerologLogger)
	sessionstoreConfig := sessionstore.ProvideConfig()
	sessionstoreService := sessionstore.ProvideService(zerologLogger, sessionstoreConfig, registry)
	getNoscriptEventsCustom := handlers.ProvideGetNoscriptEventsCustom(eventstoreService, saltmanagerService, sessionstoreService)
	getNoscriptEventsOutboundLinks := handlers.ProvideGetNoscriptEventsOutboundLinks(eventstoreService, sessionstoreService, saltmanagerService)
	uaparserService := uaparser.ProvideService(zerologLogger, registry)
	ipgeolocatorService := ipgeolocator.ProvideMmdbService(zerologLogger, registry)
	getNoscriptEventsPageviews := handlers.ProvideGetNoscriptEventsPageviews(zerologLogger, eventstoreService, uaparserService, ipgeolocatorService, saltmanagerService, sessionstoreService)
	accessLog := middlewares.ProvideAccessLog(server, zerologLogger)
	errorHandler := middlewares.ProvideErrorHandler(registry, zerologLogger)
	fiberConfig := wired.ProvideMinimalFiberConfig(server)
	healhCheck := handlers.ProvideHealthCheck()
	requestId := middlewares.ProvideRequestId(server)
	static := middlewares.ProvideStatic(server)
	metrics := middlewares.ProvideMetrics(registry)
	minimalFiber := wired.ProvideMinimalFiber(accessLog, errorHandler, fiberConfig, healhCheck, zerologLogger, requestId, static, metrics, service)
	originregistryService := originregistry.ProvideEnvVarService(zerologLogger)
	nonRegisteredOriginFilter := middlewares.ProvideNonRegisteredOriginFilter(originregistryService)
	noscriptHandlersCache := middlewares.ProvideNoscriptHandlersCache()
	postEventsCustom := handlers.ProvidePostEventsCustom(eventstoreService, saltmanagerService, sessionstoreService)
	postEventsFileDownloads := handlers.ProvidePostEventsFileDownloads(eventstoreService, saltmanagerService, sessionstoreService)
	postEventsOutboundLinks := handlers.ProvidePostEventsOutboundLinks(eventstoreService, saltmanagerService, sessionstoreService)
	postEventsPageviews := handlers.ProvidePostEventsPageViews(zerologLogger, eventstoreService, uaparserService, ipgeolocatorService, saltmanagerService, sessionstoreService)
	referrerAsDefaultOrigin := middlewares.ProvideReferrerAsDefaultOrigin()
	app := wired.ProvideFiber(apiEventsTimeout, eventsCors, eventsRateLimiter, getNoscriptEventsCustom, getNoscriptEventsOutboundLinks, getNoscriptEventsPageviews, minimalFiber, nonRegisteredOriginFilter, noscriptHandlersCache, postEventsCustom, postEventsFileDownloads, postEventsOutboundLinks, postEventsPageviews, referrerAsDefaultOrigin)
	promhttpLogger := wired.ProvidePromHttpLogger(server, zerologLogger)
	grafanaConfig := grafana.ProvideConfig(zerologLogger)
	client := grafana.ProvideClient(grafanaConfig)
	clickhouseConfig := clickhouse.ProvideConfig(zerologLogger)
	grafanaService := grafana2.ProvideService(client, clickhouseConfig)
	setup := ProvideSetup(zerologLogger, client, grafanaService)
	wiredApp := wired.ProvideApp(app, server, zerologLogger, promhttpLogger, registry, setup, service)
	return wiredApp
}
