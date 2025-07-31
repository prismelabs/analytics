package eventstore

import (
	"errors"
	"time"

	"github.com/negrel/configue"
)

// Config holds service configuration.
type Config struct {
	MaxBatchSize      uint64
	MaxBatchTimeout   time.Duration
	RingBuffersFactor uint64
}

// RegisterOptions registers Config fields as options.
func (c *Config) RegisterOptions(f *configue.Figue) {
	f.Uint64Var(&c.MaxBatchSize, "eventstore.max.batch.size", 4096, "maximum `size` of an event's batch")
	f.DurationVar(&c.MaxBatchTimeout, "eventstore.max.batch.timeout", 1*time.Minute, "maximum `duration` before a batch is sent")
	f.Uint64Var(&c.RingBuffersFactor, "eventstore.ring.buffers.factor", 100, "events ring buffer `size`")
}

// Validate validates configuration options.
func (c *Config) Validate() error {
	var errs []error
	if c.MaxBatchSize < 1 {
		errs = append(errs, errors.New("event store maximum batch size must be greater than or equal to 1"))
	}
	if c.MaxBatchTimeout < time.Second {
		errs = append(errs, errors.New("event store maximum batch timeout must be greater than or equal to 1s"))
	}
	if c.RingBuffersFactor < 1 {
		errs = append(errs, errors.New("event store ring buffer factor must be greater than or equal to 1"))
	}
	return errors.Join(errs...)
}
