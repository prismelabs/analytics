package sessionstorage

import (
	"time"

	"github.com/prismelabs/analytics/pkg/config"
)

// Session storage service configuration options.
type Config struct {
	gcInterval         time.Duration
	sessionInactiveTtl time.Duration
}

// ProvideConfig is a wire provider for session storage configuration.
func ProvideConfig() Config {
	return Config{
		gcInterval:         config.ParseDurationEnvOrDefault("PRISME_SESSIONSTORAGE_GC_INTERVAL", 10*time.Minute),
		sessionInactiveTtl: config.ParseDurationEnvOrDefault("PRISME_SESSIONSTORAGE_SESSION_INACTIVE_TTL", 24*time.Hour),
	}
}
