package options

import (
	"errors"
	"math"
	"time"

	"github.com/negrel/configue"
)

// HTTP Server options.
type Server struct {
	// Access log file path.
	AccessLog string
	// Sets log level to debug.
	Debug bool
	// Listening port.
	Port uint
	// Timeout for /api/v1/events/* handlers.
	ApiEventsTimeout time.Duration
	// Comma separated list of origins that may access /api/v1/stats/* resources.
	ApiStatsAllowOrigins string
}

// RegisterOptions registers options in provided Figue.
func (s *Server) RegisterOptions(f *configue.Figue) {
	f.StringVar(&s.AccessLog, "server.access.log", "/dev/stdout", "`filepath` to access log")
	f.BoolVar(&s.Debug, "server.debug", false, "enable debug log")
	f.UintVar(&s.Port, "server.port", 80, "HTTP server port to listen on")
	f.DurationVar(&s.ApiEventsTimeout, "server.api.events.timeout", 3*time.Second, "`duration` before handlers /api/*/events/* timeout")
	f.StringVar(&s.ApiStatsAllowOrigins, "server.api.stats.allow.origins", "", "comma separated list of `origins` that may access /api/*/stats/* resources")
}

// Validate validates configuration options.
func (s *Server) Validate() error {
	var errs []error
	if s.ApiEventsTimeout <= 0 {
		errs = append(errs, errors.New("invalid timeout option for /api/* events handlers"))
	}
	if s.Port > math.MaxUint16 {
		errs = append(errs, errors.New("invalid port for HTTP server"))
	}

	return errors.Join(errs...)
}
