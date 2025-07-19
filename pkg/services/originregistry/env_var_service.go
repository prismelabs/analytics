package originregistry

import (
	"context"
	"errors"
	"strings"

	"github.com/negrel/configue"
	"github.com/prismelabs/analytics/pkg/log"
)

type service struct {
	logger  log.Logger
	origins map[string]struct{}
}

// NewService returns a new origin registry Service.
func NewService(cfg Config, logger log.Logger) Service {
	logger = logger.With(
		"service", "originregistry",
		"service_impl", "envvar",
	)

	origins := make(map[string]struct{})
	for _, src := range strings.Split(cfg.Origins, ",") {
		origins[src] = struct{}{}
	}

	logger.Info("env var based origin registry configured", "origins", origins)

	return service{logger, origins}
}

// IsOriginRegistered implements Service.
func (evs service) IsOriginRegistered(_ context.Context, origin string) (bool, error) {
	_, ok := evs.origins[origin]
	evs.logger.Debug(
		"checked if origin is registered",
		"origin", origin,
		"origin_registered", ok,
	)

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
