package originregistry

import (
	"context"
	"errors"
	"strings"

	"github.com/negrel/configue"
	"github.com/rs/zerolog"
)

type EnvVarService struct {
	zerolog.Logger
	origins map[string]struct{}
}

// ProvideEnvVarService is a wire provider for origin registry Service.
func ProvideEnvVarService(cfg Config, logger zerolog.Logger) Service {
	logger = logger.With().
		Str("service", "originregistry").
		Str("service_impl", "envvar").
		Logger()

	origins := make(map[string]struct{})
	for _, src := range strings.Split(cfg.Origins, ",") {
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

// Service options.
type Config struct {
	Origins string
}

// RegisterOptions registers Config fields as options.
func (c *Config) RegisterOptions(f *configue.Figue) {
	f.StringVar(&c.Origins, "origin.registry.origins", "", "comma separatel `list` (without whitespace) of valid origins. Events from unknown origins are rejected")
}

// Validate validates configuration options.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.Origins) == "" {
		return errors.New("origin registry origin list is empty")
	}
	return nil
}
