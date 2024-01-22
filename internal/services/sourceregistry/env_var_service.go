package sourceregistry

import (
	"context"
	"os"
	"strings"

	"github.com/prismelabs/prismeanalytics/internal/log"
)

type EnvVarService struct {
	sources map[string]struct{}
}

// ProvideEnvVarService is a wire provider for source registry Service.
func ProvideEnvVarService(logger log.Logger) Service {
	rawSources, ok := os.LookupEnv("PRISME_SOURCE_REGISTRY_SOURCES")
	if !ok {
		panic("PRISME_SOURCE_REGISTRY_SOURCES environment variable not set")
	}

	sources := make(map[string]struct{})

	for _, src := range strings.Split(rawSources, ",") {
		sources[src] = struct{}{}
	}

	logger.Info().Any("sources", sources).Msg("env var based source registry configured")

	return EnvVarService{sources}
}

// IsSourceRegistered implements Service.
func (evs EnvVarService) IsSourceRegistered(_ context.Context, src Source) (bool, error) {
	_, ok := evs.sources[src.SourceString()]
	return ok, nil
}
