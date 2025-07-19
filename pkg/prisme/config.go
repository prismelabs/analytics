package prisme

import (
	"errors"
	"math"
	"net"
	"time"

	"github.com/negrel/configue"
)

// Config define Prisme server configuration.
type Config struct {
	// Access log file path.
	AccessLog string
	// Sets log level to debug.
	Debug bool
	// Listening port.
	Port uint
	// Trust proxy headers.
	TrustProxy bool
	// X-Forwarded-For proxy header.
	ProxyHeader string
	// X-Request-Id proxy header.
	ProxyRequestIdHeader string
	// host:port address of admin http server.
	AdminHostPort string
	// Timeout for /api/v1/events/* handlers.
	ApiEventsTimeout time.Duration
}

// RegisterOptions registers Config fields as options.
func (c *Config) RegisterOptions(f *configue.Figue) {
	f.StringVar(&c.AccessLog, "access.log", "/dev/stdout", "`filepath` to access log")
	f.BoolVar(&c.Debug, "debug", false, "enable debug log")
	f.UintVar(&c.Port, "port", 80, "HTTP server port to listen on")
	f.BoolVar(&c.TrustProxy, "trust.proxy", false, "trust proxy headers (X-Forwarded-For, X-Request-Id)")
	f.StringVar(&c.ProxyHeader, "proxy.header", "X-Forwarded-For", "HTTP header used to determine client IP address.")
	f.StringVar(&c.ProxyRequestIdHeader, "proxy.request.id.header", "X-Request-Id", "HTTP header used to retrieve request ID")
	f.StringVar(&c.AdminHostPort, "admin.hostport", "127.0.0.1:9090", "use `host:port` for administration HTTP server")
	f.DurationVar(&c.ApiEventsTimeout, "api.events.timeout", 3*time.Second, "`duration` before handlers under /api/* timeout")
}

// Validate validates configuration options.
func (c *Config) Validate() error {
	var errs []error
	_, _, err := net.SplitHostPort(c.AdminHostPort)
	if err != nil {
		errs = append(errs, errors.New("invalid administration server hostport"))
	}
	if c.ApiEventsTimeout <= 0 {
		errs = append(errs, errors.New("invalid timeout option for /api/* events handlers"))
	}
	if c.Port > math.MaxUint16 {
		errs = append(errs, errors.New("invalid port for HTTP server"))
	}

	return errors.Join(errs...)
}
