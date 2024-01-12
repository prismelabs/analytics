//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/handlers"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
	"github.com/prismelabs/prismeanalytics/internal/postgres"
	"github.com/prismelabs/prismeanalytics/internal/services/users"
)

func initialize(logger BootstrapLogger) App {
	wire.Build(
		ProvideConfig,
		wire.FieldsOf(new(config.Config), "Server"),
		wire.FieldsOf(new(config.Config), "Postgres"),
		postgres.ProvidePg,
		users.ProvideStore,
		users.ProvideService,
		ProvideLogger,
		ProvideFiberViewsEngine,
		middlewares.ProvideStatic,
		middlewares.ProvideRequestId,
		middlewares.ProvideAccessLog,
		middlewares.ProvideLogger,
		handlers.ProvideGetSignUp,
		handlers.ProvidePostSignUp,
		ProvideFiber,
		ProvideApp,
	)
	return App{}
}
