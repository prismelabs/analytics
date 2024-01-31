//go:build wireinject
// +build wireinject

package full

import (
	"github.com/google/wire"
	"github.com/prismelabs/prismeanalytics/cmd/server/wired"
	"github.com/prismelabs/prismeanalytics/internal/clickhouse"
	grafanaCli "github.com/prismelabs/prismeanalytics/internal/grafana"
	"github.com/prismelabs/prismeanalytics/internal/handlers"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
	"github.com/prismelabs/prismeanalytics/internal/services/eventstore"
	"github.com/prismelabs/prismeanalytics/internal/services/grafana"
	"github.com/prismelabs/prismeanalytics/internal/services/ipgeolocator"
	"github.com/prismelabs/prismeanalytics/internal/services/sourceregistry"
	"github.com/prismelabs/prismeanalytics/internal/services/uaparser"
)

func Initialize(logger wired.BootstrapLogger) wired.App {
	wire.Build(
		ProvideFiber,
		ProvideSetup,
		clickhouse.ProvideCh,
		eventstore.ProvideClickhouseService,
		grafana.ProvideService,
		grafanaCli.ProvideClient,
		handlers.ProvideHealthCheck,
		handlers.ProvidePostEventsPageViews,
		ipgeolocator.ProvideMmdbService,
		middlewares.ProvideAccessLog,
		middlewares.ProvideEventsCors,
		middlewares.ProvideEventsRateLimiter,
		middlewares.ProvideLogger,
		middlewares.ProvideRequestId,
		middlewares.ProvideStatic,
		sourceregistry.ProvideEnvVarService,
		uaparser.ProvideService,
		wired.ProvideApp,
		wired.ProvideClickhouseConfig,
		wired.ProvideGrafanaConfig,
		wired.ProvideLogger,
		wired.ProvideMinimalFiber,
		wired.ProvideServerConfig,
	)
	return wired.App{}
}
