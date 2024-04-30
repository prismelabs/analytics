package eventstore

import (
	"time"

	"github.com/prismelabs/analytics/pkg/config"
)

// Config holds service configuration.
type Config struct {
	MaxBatchSize    uint64
	MaxBatchTimeout time.Duration
}

// ProvideConfig is a wire provider for service config.
func ProvideConfig() Config {
	maxBatchSize := config.ParseUintEnvOrDefault("PRISME_EVENTSTORE_MAX_BATCH_SIZE", 4096, 64)
	maxBatchTimeout := config.ParseDurationEnvOrDefault("PRISME_EVENTSTORE_MAX_BATCH_TIMEOUT", 1*time.Minute)

	return Config{
		MaxBatchSize:    maxBatchSize,
		MaxBatchTimeout: maxBatchTimeout,
	}
}
