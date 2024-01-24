//go:build wireinject
// +build wireinject

package ingestion

import (
	"github.com/google/wire"
	"github.com/prismelabs/prismeanalytics/cmd/server/wired"
	"github.com/prismelabs/prismeanalytics/internal/clickhouse"
	"github.com/prismelabs/prismeanalytics/internal/handlers"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
	"github.com/prismelabs/prismeanalytics/internal/services/eventstore"
	"github.com/prismelabs/prismeanalytics/internal/services/sourceregistry"
	"github.com/prismelabs/prismeanalytics/internal/services/uaparser"
)

func Initialize(logger wired.BootstrapLogger) wired.App {
	wire.Build(
		wired.ProvideServerConfig,
		wired.ProvideClickhouseConfig,
		wired.ProvideLogger,
		ProvideFiber,
		wired.ProvideApp,
		wired.ProvideFiberViewsEngine, // not used.
		wired.ProvideMinimalFiber,
		middlewares.ProvideLogger,
		middlewares.ProvideStatic,
		middlewares.ProvideAccessLog,
		middlewares.ProvideRequestId,
		middlewares.ProvideEventsCors,
		middlewares.ProvideEventsRateLimiter,
		handlers.ProvidePostEventsPageViews,
		eventstore.ProvideClickhouseService,
		sourceregistry.ProvideEnvVarService,
		clickhouse.ProvideCh,
		uaparser.ProvideService,
	)
	return wired.App{}
}
