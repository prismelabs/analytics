package originregistry

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/negrel/configue"
	"github.com/prismelabs/analytics/pkg/log"
)

// Service define an origin registry management service.
type Service interface {
	IsOriginRegistered(context.Context, string) (bool, error)
}

type service struct {
	logger      log.Logger
	origins     map[string]struct{}
	hasWildcard bool
}

// NewService returns a new origin registry Service.
func NewService(cfg Config, logger log.Logger) (Service, error) {
	logger = logger.With(
		"service", "originregistry",
		"service_impl", "envvar",
	)

	srv := &service{logger: logger,
		origins:     make(map[string]struct{}),
		hasWildcard: false,
	}

	origins := make(map[string]struct{})
	for _, origin := range cfg.Origins {
		wildcard := false

		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}

		if origin[0] == '*' {
			// *.negrel.dev become .negrel.dev
			origin = origin[1:]
			wildcard = true

			// *.fr and *bar.foo.fr are not allowed.
			if strings.Count(origin, ".") < 2 || !strings.HasPrefix(origin, ".") {
				return nil, errors.New("wildcard is only allowed at subdomain level (e.g. *.negrel.dev or *.www.negrel.dev)")
			}
		}

		// foo..fr and .fr are invalid.
		labels := strings.Split(origin, ".")
		for i, l := range labels {
			if (i > 0 || !wildcard) && strings.TrimSpace(l) == "" {
				return nil, fmt.Errorf("invalid origin %q", origin)
			}
		}

		// www.*.negrel.dev is not allowed.
		if strings.ContainsRune(origin, '*') {
			return nil, errors.New("wildcard is only allowed at the beginning of an origin (e.g. *.negrel.dev)")
		}

		srv.hasWildcard = srv.hasWildcard || wildcard
		srv.origins[origin] = struct{}{}
	}

	if len(srv.origins) == 0 {
		return nil, errors.New("no valid origin provided")
	}

	logger.Info("env var based origin registry configured", "origins", origins)

	return srv, nil
}

// IsOriginRegistered implements Service.
func (evs *service) IsOriginRegistered(_ context.Context, origin string) (bool, error) {
	if len(origin) > 256 {
		return false, nil
	}

	_, ok := evs.origins[origin]
	if !ok && evs.hasWildcard {
		// if origin is foo.bar.negrel.dev we first check
		// if there is .negrel.dev then .bar.negrel.dev
		// .negrel.dev should match *.negrel.dev
		tldIdx := strings.LastIndexByte(origin, '.')
		if tldIdx < 0 {
			goto end
		}
		domainIdx := strings.LastIndexByte(origin[:tldIdx], '.')
		if domainIdx < 0 {
			goto end
		}

		idx := domainIdx
		for idx != -1 {
			_, ok = evs.origins[origin[idx:]]
			if ok {
				goto end
			}
			idx = strings.LastIndexByte(origin[:idx], '.')
		}
	}

end:
	evs.logger.Debug(
		"checked if origin is registered",
		"origin", origin,
		"origin_registered", ok,
	)
	return ok, nil
}

// Service options.
type Config struct {
	Origins []string
}

// RegisterOptions registers Config fields as options.
func (c *Config) RegisterOptions(f *configue.Figue) {
	f.StringSliceVar(&c.Origins, "origins", nil, "comma separated `list` of allowed origins without scheme (e.g. localhost, example.com, prismeanalytics.com)")
}

// Validate validates configuration options.
func (c *Config) Validate() error {
	if c.Origins == nil {
		return errors.New("origins allow list is empty, please specify -origins flag")
	}
	return nil
}
