//go:build wireinject
// +build wireinject

package full

import (
	"github.com/google/wire"
	"github.com/prismelabs/prismeanalytics/cmd/server/wired"
	"github.com/prismelabs/prismeanalytics/pkg/clickhouse"
	grafanaCli "github.com/prismelabs/prismeanalytics/pkg/grafana"
	"github.com/prismelabs/prismeanalytics/pkg/handlers"
	"github.com/prismelabs/prismeanalytics/pkg/middlewares"
	"github.com/prismelabs/prismeanalytics/pkg/services/eventstore"
	"github.com/prismelabs/prismeanalytics/pkg/services/grafana"
	"github.com/prismelabs/prismeanalytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/prismeanalytics/pkg/services/sourceregistry"
	"github.com/prismelabs/prismeanalytics/pkg/services/uaparser"
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
		middlewares.ProvideErrorHandler,
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
