//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/prismelabs/prismeanalytics/internal/log"
	"github.com/prismelabs/prismeanalytics/internal/renderer"
)

func initialize(logger log.Logger) App {
	wire.Build(
		ProvideConfig,
		ProvideStandardLogger,
		ProvideAccessLogger,
		renderer.ProvideRenderer,
		ProvideFiber,
		ProvideApp,
	)
	return App{}
}
