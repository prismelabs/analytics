package eventstore

import (
	"time"

	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/rs/zerolog"
)

// Config holds service configuration.
type Config struct {
	Backend           string
	BackendConfig     any
	MaxBatchSize      uint64
	MaxBatchTimeout   time.Duration
	RingBuffersFactor uint64
}

// ProvideConfig is a wire provider for service config.
func ProvideConfig(logger zerolog.Logger) Config {
	backend := config.GetEnvOrDefault("PRISME_EVENTSTORE_BACKEND", "clickhouse")

	var backendConfig any
	switch backend {
	case "clickhouse":
		backendConfig = clickhouse.ProvideConfig(logger)
	case "chdb":
		backendConfig = chdb.ProvideConfig(logger)
	default:
		logger.Panic().Msgf("invalid event store backend (%v), must be one of 'clickhouse', 'chdb'", backend)
	}

	maxBatchSize := config.ParseUintEnvOrDefault("PRISME_EVENTSTORE_MAX_BATCH_SIZE", 4096, 64)
	maxBatchTimeout := config.ParseDurationEnvOrDefault("PRISME_EVENTSTORE_MAX_BATCH_TIMEOUT", 1*time.Minute)
	ringBufferFactor := config.ParseUintEnvOrDefault("PRISME_EVENTSTORE_RING_BUFFERS_FACTOR", 100, 64)

	return Config{
		Backend:           backend,
		BackendConfig:     backendConfig,
		MaxBatchSize:      maxBatchSize,
		MaxBatchTimeout:   maxBatchTimeout,
		RingBuffersFactor: ringBufferFactor,
	}
}
