package eventstore

import (
	"time"

	"github.com/prismelabs/analytics/pkg/config"
)

// Config holds service configuration.
type Config struct {
	MaxBatchSize      uint64
	MaxBatchTimeout   time.Duration
	RingBuffersFactor uint64
}

// ProvideConfig is a wire provider for service config.
func ProvideConfig() Config {
	maxBatchSize := config.ParseUintEnvOrDefault("PRISME_EVENTSTORE_MAX_BATCH_SIZE", 4096, 64)
	maxBatchTimeout := config.ParseDurationEnvOrDefault("PRISME_EVENTSTORE_MAX_BATCH_TIMEOUT", 1*time.Minute)
	ringBufferFactor := config.ParseUintEnvOrDefault("PRISME_EVENTSTORE_RING_BUFFERS_FACTOR", 100, 64)

	return Config{
		MaxBatchSize:      maxBatchSize,
		MaxBatchTimeout:   maxBatchTimeout,
		RingBuffersFactor: ringBufferFactor,
	}
}
