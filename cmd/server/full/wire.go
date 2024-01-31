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
	"github.com/prismelabs/prismeanalytics/internal/postgres"
	"github.com/prismelabs/prismeanalytics/internal/services/auth"
	"github.com/prismelabs/prismeanalytics/internal/services/eventstore"
	"github.com/prismelabs/prismeanalytics/internal/services/grafana"
	"github.com/prismelabs/prismeanalytics/internal/services/ipgeolocator"
	"github.com/prismelabs/prismeanalytics/internal/services/sessions"
	"github.com/prismelabs/prismeanalytics/internal/services/sourceregistry"
	"github.com/prismelabs/prismeanalytics/internal/services/uaparser"
	"github.com/prismelabs/prismeanalytics/internal/services/users"
)

func Initialize(logger wired.BootstrapLogger) wired.App {
	wire.Build(
		ProvideFiber,
		ProvideSetup,
		auth.ProvideService,
		clickhouse.ProvideCh,
		eventstore.ProvideClickhouseService,
		grafana.ProvideService,
		grafanaCli.ProvideClient,
		handlers.ProvideGetIndex,
		handlers.ProvideGetSignIn,
		handlers.ProvideGetSignUp,
		handlers.ProvideNotFound,
		handlers.ProvidePostEventsPageViews,
		handlers.ProvidePostSignIn,
		handlers.ProvidePostSignUp,
		ipgeolocator.ProvideMmdbService,
		middlewares.ProvideAccessLog,
		middlewares.ProvideEventsCors,
		middlewares.ProvideEventsRateLimiter,
		middlewares.ProvideFavicon,
		middlewares.ProvideLogger,
		middlewares.ProvideRequestId,
		middlewares.ProvideStatic,
		middlewares.ProvideWithSession,
		postgres.ProvidePg,
		sessions.ProvideService,
		sourceregistry.ProvideEnvVarService,
		uaparser.ProvideService,
		users.ProvideService,
		wired.ProvideApp,
		wired.ProvideClickhouseConfig,
		wired.ProvideFiberViewsEngine,
		wired.ProvideGrafanaConfig,
		wired.ProvideLogger,
		wired.ProvideMinimalFiber,
		wired.ProvidePostgresConfig,
		wired.ProvideServerConfig,
	)
	return wired.App{}
}
