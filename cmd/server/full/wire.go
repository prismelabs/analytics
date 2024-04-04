//go:build wireinject
// +build wireinject

package full

import (
	"github.com/google/wire"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	grafanaCli "github.com/prismelabs/analytics/pkg/grafana"
	"github.com/prismelabs/analytics/pkg/handlers"
	"github.com/prismelabs/analytics/pkg/middlewares"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/grafana"
	"github.com/prismelabs/analytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/analytics/pkg/services/originregistry"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prismelabs/analytics/pkg/wired"
)

func Initialize(logger wired.BootstrapLogger) wired.App {
	wire.Build(
		ProvideFiber,
		ProvideSetup,
		clickhouse.ProvideCh,
		clickhouse.ProvideEmbeddedSourceDriver,
		eventstore.ProvideClickhouseService,
		grafana.ProvideService,
		grafanaCli.ProvideClient,
		handlers.ProvideHealthCheck,
		handlers.ProvidePostEventsCustom,
		handlers.ProvidePostEventsPageViews,
		ipgeolocator.ProvideMmdbService,
		middlewares.ProvideAccessLog,
		middlewares.ProvideErrorHandler,
		middlewares.ProvideEventsCors,
		middlewares.ProvideEventsRateLimiter,
		middlewares.ProvideNonRegisteredOriginFilter,
		middlewares.ProvideRequestId,
		middlewares.ProvideStatic,
		originregistry.ProvideEnvVarService,
		teardown.ProvideService,
		uaparser.ProvideService,
		wired.ProvideApp,
		wired.ProvideClickhouseConfig,
		wired.ProvideGrafanaConfig,
		wired.ProvideLogger,
		wired.ProvideMinimalFiber,
		wired.ProvideMinimalFiberConfig,
		wired.ProvideServerConfig,
	)
	return wired.App{}
}
