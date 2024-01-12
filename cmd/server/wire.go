//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
)

func initialize(logger BootstrapLogger) App {
	wire.Build(
		ProvideConfig,
		wire.FieldsOf(new(config.Config), "Server"),
		ProvideLogger,
		ProvideFiberViewsEngine,
		middlewares.ProvideStatic,
		middlewares.ProvideRequestId,
		middlewares.ProvideAccessLog,
		middlewares.ProvideLogger,
		ProvideFiber,
		ProvideApp,
	)
	return App{}
}
