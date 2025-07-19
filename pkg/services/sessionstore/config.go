package sessionstore

import (
	"errors"
	"time"

	"github.com/negrel/configue"
)

// Session storage service configuration options.
type Config struct {
	gcInterval             time.Duration
	sessionInactiveTtl     time.Duration
	deviceExpiryPercentile int
	maxSessionsPerVisitor  uint64
}

// RegisterOptions registers Config fields as options.
func (c *Config) RegisterOptions(f *configue.Figue) {
	f.DurationVar(&c.gcInterval, "sessionstore.gc.interval", 10*time.Second, "interval at which expired sessions are collected")
	f.DurationVar(&c.sessionInactiveTtl, "sessionstore.session.inactive.ttl", 24*time.Hour, "`duration` before inactive session expires")
	f.IntVar(&c.deviceExpiryPercentile, "sessionstore.device.expiry.percentile", 50, "minimum percentage of expired sessions triggering device session cleanup")
	f.Uint64Var(&c.maxSessionsPerVisitor, "sessionstore.max.sessions.per.visitor", 64, "maximum number of sessions per visitor/device")
}

// Validate validates configuration options.
func (c *Config) Validate() error {
	var errs []error
	if c.gcInterval < 1*time.Second {
		errs = append(errs, errors.New("sessionstore gc interval must be greater than 1s"))
	}
	if c.sessionInactiveTtl < 1*time.Second {
		errs = append(errs, errors.New("sessionstore inactive session TTL must be greater than 1s"))
	}
	if c.deviceExpiryPercentile <= 0 {
		errs = append(errs, errors.New("sessionstore device expiry percentile must be greater than 0"))
	}
	if c.maxSessionsPerVisitor <= 0 {
		errs = append(errs, errors.New("sessionstore max session per visitor must be greater than 0"))
	}
	return errors.Join(errs...)
}
