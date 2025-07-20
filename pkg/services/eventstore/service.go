package eventstore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/negrel/ringo"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/retry"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prometheus/client_golang/prometheus"
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

var backendsFactory = map[string]func(
	logger log.Logger,
	cfg any,
	source source.Driver,
	teardown teardown.Service,
) backend{}

// NewService returns a new event store service.
func NewService(
	cfg Config,
	logger log.Logger,
	promRegistry *prometheus.Registry,
	teardown teardown.Service,
	source source.Driver,
) (Service, error) {
	fact := backendsFactory[cfg.Backend]
	if fact == nil {
		return nil, fmt.Errorf("no %v event store backend", cfg.Backend)
	}

	return newService(
		cfg,
		fact(logger, cfg.BackendConfig, source, teardown),
		logger,
		promRegistry,
		teardown,
	), nil
}

type service struct {
	logger          log.Logger
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
	logger log.Logger,
	promRegistry *prometheus.Registry,
	teardownService teardown.Service,
) Service {
	batchDone := make(chan struct{})
	logger = logger.With(
		"service", "eventstore",
		"backend", cfg.Backend,
		"ring_buffers_factor", cfg.RingBuffersFactor,
		"max_batch_size", cfg.MaxBatchSize,
		"max_batch_timeout", cfg.MaxBatchTimeout,
	)

	// Create context for batch loops.
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel them on teardown.
	teardownService.RegisterProcedure(func() error {
		logger.Info("cancelling event batch loops...")
		cancel()
		// Wait for last batch to be sent.
		<-batchDone
		logger.Info("event batch loops canceled.")
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
					logger.Warn("events ring buffer collision detected, consider increasing PRISME_EVENTSTORE_RING_BUFFERS_FACTOR or PRISME_EVENTSTORE_MAX_BATCH_SIZE")
				})),
			),
			ringo.WithWaiterContext[any](ctx),
		),
		metrics: newMetrics(promRegistry),
	}

	go service.batchLoop(ctx, batchDone)

	logger.Info("event store configured", "backend", cfg.Backend)

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
					s.logger.Debug("preparing a new event batch", "try", n)
					return s.backend.prepareBatch()
				},
				retry.CancelOnContextError,
			)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					break
				}
				s.logger.Err("failed to prepare next event batch", err)
				continue
			}

			batchCreationDate = time.Now()
			s.logger.Debug("new event batch prepared", "date", batchCreationDate)
		}

		// Wait for next event.
		s.logger.Debug("waiting for event...")
		ev, done, dropped := s.eventRingBuf.Next()
		// Ring buffer context was canceled.
		if done {
			s.logger.Info("events ring buffer done, sending last batch...")
			err = s.backend.sendBatch()
			if err != nil {
				s.logger.Err("failed to send last batch of events", err)
			} else {
				s.logger.Info("last batch of events sent")
			}
			break
		}
		if dropped > 0 {
			s.logger.Info("events dropped", "dropped", dropped)
			s.metrics.droppedEvents.Add(float64(dropped))
		}

		// Append to batch.
		s.logger.Debug("appending event to batch...", "event", ev)
		err = s.backend.appendToBatch(ev)
		if err != nil {
			s.logger.Err("failed to append event to batch", err)
		} else {
			batchSize++
		}

		if uint64(batchSize) >= s.maxBatchSize || time.Since(batchCreationDate) > s.maxBatchTimeout {
			s.sendBatch(batchSize)
			batchSize = 0
		}
	}

	batchDone <- struct{}{}
	s.logger.Info("eventstore batch loop done")
}

func (s *service) sendBatch(batchSize int) {
	// Retry if an error occurred. This can happen on clickhouse cloud if instance
	// goes to idle state.
	var err error

	s.logger.Debug("sending event batch...")
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
			s.logger.Debug(
				"events batch successfully sent",
				"send_duration", dur,
				"batch_size", batchSize,
			)

			return nil
		}
	}, retry.NeverCancel)

	if err != nil {
		s.metrics.batchDropped.Inc()
		s.logger.Err("failed to send events batch", err)
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
