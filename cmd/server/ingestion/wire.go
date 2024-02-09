//go:build wireinject
// +build wireinject

package ingestion

import (
	"github.com/google/wire"
	"github.com/prismelabs/prismeanalytics/cmd/server/wired"
	"github.com/prismelabs/prismeanalytics/pkg/clickhouse"
	"github.com/prismelabs/prismeanalytics/pkg/handlers"
	"github.com/prismelabs/prismeanalytics/pkg/middlewares"
	"github.com/prismelabs/prismeanalytics/pkg/services/eventstore"
	"github.com/prismelabs/prismeanalytics/pkg/services/ipgeolocator"
	"github.com/prismelabs/prismeanalytics/pkg/services/sourceregistry"
	"github.com/prismelabs/prismeanalytics/pkg/services/uaparser"
)

func Initialize(logger wired.BootstrapLogger) wired.App {
	wire.Build(
		ProvideFiber,
		clickhouse.ProvideCh,
		eventstore.ProvideClickhouseService,
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
		wired.ProvideLogger,
		wired.ProvideMinimalFiber,
		wired.ProvideServerConfig,
		wired.ProvideSetup,
	)
	return wired.App{}
}
