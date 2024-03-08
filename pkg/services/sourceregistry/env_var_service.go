package sourceregistry

import (
	"context"
	"strings"

	"github.com/prismelabs/analytics/pkg/config"
	"github.com/rs/zerolog"
)

type EnvVarService struct {
	zerolog.Logger
	sources map[string]struct{}
}

// ProvideEnvVarService is a wire provider for source registry Service.
func ProvideEnvVarService(logger zerolog.Logger) Service {
	logger = logger.With().
		Str("service", "sourceregistry").
		Str("service_impl", "envvar").
		Logger()

	rawSources := config.MustGetEnv("PRISME_SOURCE_REGISTRY_SOURCES")

	sources := make(map[string]struct{})

	for _, src := range strings.Split(rawSources, ",") {
		sources[src] = struct{}{}
	}

	logger.Info().Any("sources", sources).Msg("env var based source registry configured")

	return EnvVarService{logger, sources}
}

// IsSourceRegistered implements Service.
func (evs EnvVarService) IsSourceRegistered(_ context.Context, src Source) (bool, error) {
	_, ok := evs.sources[src.SourceString()]
	evs.Logger.Debug().Bool("source_registered", ok).Send()

	return ok, nil
}
