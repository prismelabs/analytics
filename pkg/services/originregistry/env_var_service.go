package originregistry

import (
	"context"
	"fmt"
	"strings"

	"github.com/prismelabs/analytics/pkg/config"
	"github.com/rs/zerolog"
)

type EnvVarService struct {
	zerolog.Logger
	origins map[string]struct{}
}

// ProvideEnvVarService is a wire provider for origin registry Service.
func ProvideEnvVarService(logger zerolog.Logger) Service {
	logger = logger.With().
		Str("service", "originregistry").
		Str("service_impl", "envvar").
		Logger()

	rawOrigins := config.GetEnvOrDefault("PRISME_ORIGIN_REGISTRY_ORIGINS", "")
	if rawOrigins == "" {
		// Deprecated.
		rawOrigins = config.GetEnvOrDefault("PRISME_SOURCE_REGISTRY_SOURCES", "")
	} else {
		if config.GetEnvOrDefault("PRISME_SOURCE_REGISTRY_SOURCES", "") != "" {
			panic("PRISME_ORIGIN_REGISTRY_ORIGINS and PRISME_SOURCE_REGISTRY_SOURCES are both set but they are mutually exclusive, please only use PRISME_ORIGIN_REGISTRY_ORIGINS")
		}
	}
	if rawOrigins == "" {
		panic(fmt.Errorf("PRISME_ORIGIN_REGISTRY_ORIGINS environment variable is not set or is an empty string"))
	}

	origins := make(map[string]struct{})
	for _, src := range strings.Split(rawOrigins, ",") {
		origins[src] = struct{}{}
	}

	logger.Info().Any("origins", origins).Msg("env var based origin registry configured")

	return EnvVarService{logger, origins}
}

// IsOriginRegistered implements Service.
func (evs EnvVarService) IsOriginRegistered(_ context.Context, origin string) (bool, error) {
	_, ok := evs.origins[origin]
	evs.Logger.Debug().
		Str("origin", origin).
		Bool("origin_registered", ok).
		Msg("checked if origin is registered")

	return ok, nil
}
