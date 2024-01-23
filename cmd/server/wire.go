//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/prismelabs/prismeanalytics/internal/clickhouse"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/handlers"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
	"github.com/prismelabs/prismeanalytics/internal/postgres"
	"github.com/prismelabs/prismeanalytics/internal/services/auth"
	"github.com/prismelabs/prismeanalytics/internal/services/eventstore"
	"github.com/prismelabs/prismeanalytics/internal/services/sessions"
	"github.com/prismelabs/prismeanalytics/internal/services/sourceregistry"
	"github.com/prismelabs/prismeanalytics/internal/services/uaparser"
	"github.com/prismelabs/prismeanalytics/internal/services/users"
)

func initialize(logger BootstrapLogger) App {
	wire.Build(
		ProvideConfig,
		wire.FieldsOf(new(config.Config), "Server"),
		wire.FieldsOf(new(config.Config), "Postgres"),
		wire.FieldsOf(new(config.Config), "Clickhouse"),
		postgres.ProvidePg,
		clickhouse.ProvideCh,
		sessions.ProvideService,
		users.ProvideService,
		auth.ProvideService,
		eventstore.ProvideClickhouseService,
		sourceregistry.ProvideEnvVarService,
		uaparser.ProvideService,
		ProvideLogger,
		ProvideFiberViewsEngine,
		middlewares.ProvideStatic,
		middlewares.ProvideRequestId,
		middlewares.ProvideAccessLog,
		middlewares.ProvideLogger,
		middlewares.ProvideWithSession,
		middlewares.ProvideFavicon,
		middlewares.ProvideEventsCors,
		middlewares.ProvideEventsRateLimiter,
		handlers.ProvideGetSignUp,
		handlers.ProvidePostSignUp,
		handlers.ProvideGetSignIn,
		handlers.ProvidePostSignIn,
		handlers.ProvideGetIndex,
		handlers.ProvideNotFound,
		handlers.ProvidePostEventsPageViews,
		ProvideFiber,
		ProvideApp,
	)
	return App{}
}
