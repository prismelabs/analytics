package sessionstore

import (
	"fmt"
	"time"

	"github.com/prismelabs/analytics/pkg/config"
)

// Session storage service configuration options.
type Config struct {
	gcInterval             time.Duration
	sessionInactiveTtl     time.Duration
	deviceExpiryPercentile int
	maxSessionsPerVisitor  uint64
}

// ProvideConfig is a wire provider for session storage configuration.
func ProvideConfig() Config {
	deviceExpiryPercentile := int(config.ParseIntEnvOrDefault("PRISME_SESSIONSTORAGE_DEVICE_EXPIRY_PERCENTILE", 50, 8))
	if deviceExpiryPercentile > 100 || deviceExpiryPercentile < 0 {
		panic(fmt.Errorf("PRISME_SESSIONSTORAGE_DEVICE_EXPIRY_PERCENTILE must be comprise between 0 and 100"))
	}

	return Config{
		gcInterval:             config.ParseDurationEnvOrDefault("PRISME_SESSIONSTORAGE_GC_INTERVAL", 10*time.Second),
		sessionInactiveTtl:     config.ParseDurationEnvOrDefault("PRISME_SESSIONSTORAGE_SESSION_INACTIVE_TTL", 24*time.Hour),
		deviceExpiryPercentile: deviceExpiryPercentile,
		maxSessionsPerVisitor:  config.ParseUintEnvOrDefault("PRISME_SESSIONSTORAGE_MAX_SERSSIONS_PER_VISITOR", 64, 64),
	}
}
