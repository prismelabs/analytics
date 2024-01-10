//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/prismelabs/prismeanalytics/internal/log"
	"github.com/prismelabs/prismeanalytics/internal/middlewares"
)

func initialize(logger log.Logger) App {
	wire.Build(
		ProvideConfig,
		ProvideStandardLogger,
		ProvideAccessLogger,
		ProvideEcho,
		ProvideApp,
	)
	return App{}
}
