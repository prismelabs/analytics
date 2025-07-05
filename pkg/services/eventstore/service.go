package eventstore

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/negrel/ringo"
	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/retry"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var (
	ErrReadOnly = errors.New("query is not readonly")
)

// Service define an event storage service.
type Service interface {
	StorePageView(context.Context, *event.PageView) error
	StoreCustom(context.Context, *event.Custom) error
	StoreOutboundLinkClick(context.Context, *event.OutboundLinkClick) error
	StoreFileDownload(context.Context, *event.FileDownload) error
	// Only read-only query should be supported.
	Query(ctx context.Context, query string, args ...any) (QueryResult, error)
}

// ProvideService is a wire provider for event storage service.
func ProvideService(
	cfg Config,
	logger zerolog.Logger,
	promRegistry *prometheus.Registry,
	teardown teardown.Service,
	source source.Driver,
) Service {
	if cfg.Backend == "clickhouse" {
		ch := clickhouse.ProvideCh(logger, cfg.BackendConfig.(clickhouse.Config), source, teardown)
		backend := newClickhouseBackend(ch)
		return newService(cfg, backend, logger, promRegistry, teardown)
	} else {
		chdb := chdb.ProvideChDb(logger, cfg.BackendConfig.(chdb.Config), source, teardown)
		backend := newChDbBackend(chdb)
		return newService(cfg, backend, logger, promRegistry, teardown)
	}
}

type service struct {
	logger          zerolog.Logger
	backend         backend
	maxBatchSize    uint64
	maxBatchTimeout time.Duration
	eventRingBuf    ringo.Waiter[any]
	metrics         metrics
}

type backend interface {
	prepareBatch() error
	appendToBatch(any) error
	sendBatch() error
	query(ctx context.Context, query string, args ...any) (QueryResult, error)
}

func newService(
	cfg Config,
	backend backend,
	logger zerolog.Logger,
	promRegistry *prometheus.Registry,
	teardownService teardown.Service,
) Service {
	batchDone := make(chan struct{})
	logger = logger.With().
		Str("service", "eventstore").
		Str("backend", cfg.Backend).
		Uint64("ring_buffers_factor", cfg.RingBuffersFactor).
		Uint64("max_batch_size", cfg.MaxBatchSize).
		Stringer("max_batch_timeout", cfg.MaxBatchTimeout).
		Logger()

	// Create context for batch loops.
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel them on teardown.
	teardownService.RegisterProcedure(func() error {
		logger.Info().Msg("cancelling event batch loops...")
		cancel()
		// Wait for last batch to be sent.
		<-batchDone
		logger.Info().Msg("event batch loops canceled.")
		return nil
	})

	service := &service{
		logger:          logger,
		backend:         backend,
		maxBatchSize:    cfg.MaxBatchSize,
		maxBatchTimeout: cfg.MaxBatchTimeout,
		eventRingBuf: ringo.NewWaiter(
			ringo.NewManyToOne(
				int(cfg.MaxBatchSize*cfg.RingBuffersFactor),
				ringo.WithManyToOneCollisionHandler[any](ringo.CollisionHandlerFunc(func(_ any) {
					logger.Warn().Msg("events ring buffer collision detected, consider increasing PRISME_EVENTSTORE_RING_BUFFERS_FACTOR or PRISME_EVENTSTORE_MAX_BATCH_SIZE")
				})),
			),
			ringo.WithWaiterContext[any](ctx),
		),
		metrics: newMetrics(promRegistry),
	}

	go service.batchLoop(ctx, batchDone)

	logger.Info().Msgf("%v based event store configured", cfg.Backend)

	return service
}

// StoreFileDownload implements Service.
func (s *service) StoreFileDownload(_ context.Context, ev *event.FileDownload) error {
	s.eventRingBuf.Push(ev)
	return nil
}

// StoreOutboundLinkClick implements Service.
func (s *service) StoreOutboundLinkClick(_ context.Context, ev *event.OutboundLinkClick) error {
	s.eventRingBuf.Push(ev)
	return nil
}

// StorePageView implements Service.
func (s *service) StorePageView(_ context.Context, ev *event.PageView) error {
	s.eventRingBuf.Push(ev)
	return nil
}

// StoreCustom implements Service.
func (s *service) StoreCustom(_ context.Context, ev *event.Custom) error {
	s.eventRingBuf.Push(ev)
	return nil
}

func (s *service) batchLoop(ctx context.Context, batchDone chan<- struct{}) {
	var err error
	var batchSize int
	batchCreationDate := time.Now()

	for {
		if batchSize == 0 {
			err = retry.LinearRandomBackoff(5, time.Second,
				func(n uint) error {
					s.logger.Debug().Uint("try", n).Msg("preparing a new event batch")
					return s.backend.prepareBatch()
				},
				retry.CancelOnContextError,
			)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					break
				}
				s.logger.Err(err).Msg("failed to prepare next event batch")
				continue
			}

			batchCreationDate = time.Now()
			s.logger.Debug().Time("date", batchCreationDate).Msg("new event batch prepared")
		}

		// Wait for next event.
		s.logger.Debug().Msg("waiting for event...")
		ev, done, dropped := s.eventRingBuf.Next()
		// Ring buffer context was canceled.
		if done {
			s.logger.Info().Msg("events ring buffer done, sending last batch...")
			err = s.backend.sendBatch()
			if err != nil {
				s.logger.Error().Err(err).Msg("failed to send last batch of events")
			} else {
				s.logger.Info().Msg("last batch of events sent")
			}
			break
		}
		if dropped > 0 {
			s.logger.Info().Int("dropped", dropped).Msg("events dropped")
			s.metrics.droppedEvents.Add(float64(dropped))
		}

		// Append to batch.
		s.logger.Debug().Any("event", ev).Msg("appending event to batch...")
		err = s.backend.appendToBatch(ev)
		if err != nil {
			s.logger.Err(err).Msg("failed to append event to batch")
		} else {
			batchSize++
		}

		if uint64(batchSize) >= s.maxBatchSize || time.Since(batchCreationDate) > s.maxBatchTimeout {
			s.sendBatch(batchSize)
			batchSize = 0
		}
	}

	batchDone <- struct{}{}
	s.logger.Info().Msg("eventstore batch loop done")
}

func (s *service) sendBatch(batchSize int) {
	// Retry if an error occurred. This can happen on clickhouse cloud if instance
	// goes to idle state.
	var err error

	s.logger.Debug().Msg("sending event batch...")
	err = retry.LinearBackoff(5, time.Second, func(_ uint) error {
		start := time.Now()
		err := s.backend.sendBatch()

		if err != nil {
			s.metrics.batchRetry.Inc()
			return err
		} else {
			dur := time.Since(start)
			s.metrics.sendBatchDuration.Observe(dur.Seconds())
			s.metrics.batchSize.Observe(float64(batchSize))
			s.metrics.eventsCounter.Add(float64(batchSize))
			s.logger.Debug().
				Dur("send_duration", dur).
				Int("batch_size", batchSize).
				Msg("events batch successfully sent")
			return nil
		}
	}, retry.NeverCancel)

	if err != nil {
		s.metrics.batchDropped.Inc()
		s.logger.Err(err).Msg("failed to send events batch")
	}
}

// Query implements Service.
func (s *service) Query(ctx context.Context, query string, args ...any) (QueryResult, error) {
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(query)), "select ") {
		return nil, ErrReadOnly
	}

	return s.backend.query(ctx, query, args...)
}

// QueryResult define result of an event store query.
type QueryResult interface {
	Next() bool
	Scan(...any) error
	Close() error
}
