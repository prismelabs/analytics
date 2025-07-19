package eventstore

import (
	"errors"
	"time"

	"github.com/negrel/configue"
)

// Config holds service configuration.
type Config struct {
	Backend           string
	BackendConfig     any
	MaxBatchSize      uint64
	MaxBatchTimeout   time.Duration
	RingBuffersFactor uint64
}

// RegisterOptions registers Config fields as options.
func (c *Config) RegisterOptions(f *configue.Figue) {
	f.StringVar(&c.Backend, "eventstore.backend", "clickhouse", "event store `backend` ('clickhouse'/'chdb') to use")
	f.Uint64Var(&c.MaxBatchSize, "eventstore.max.batch.size", 4096, "maximum `size` of an event's batch")
	f.DurationVar(&c.MaxBatchTimeout, "eventstore.max.batch.timeout", 1*time.Minute, "maximum `duration` before a batch is sent")
	f.Uint64Var(&c.RingBuffersFactor, "eventstore.ring.buffers.factor", 100, "events ring buffer `size`")
}

// Validate validates configuration options.
func (c *Config) Validate() error {
	var errs []error
	if c.Backend != "clickhouse" && c.Backend != "chdb" {
		errs = append(errs, errors.New("event store backend must be 'clickhouse' or 'chdb'"))
	}
	if c.MaxBatchTimeout < time.Second {
		errs = append(errs, errors.New("event store maximum batch timeout must be greater than or equal to 1s"))
	}
	return errors.Join(errs...)
}
