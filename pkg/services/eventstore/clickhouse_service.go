package eventstore

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/negrel/ringo"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/rs/zerolog"
)

// ProvideClickhouseService is a wire provider for a clickhouse based event
// storage service.
func ProvideClickhouseService(ch clickhouse.Ch, logger zerolog.Logger) Service {
	service := &ClickhouseService{
		logger:          logger,
		conn:            ch.Conn,
		maxBatchSize:    config.ParseUintEnvOrDefault("PRISME_EVENTSTORE_MAX_BATCH_SIZE", 4096, 64),
		maxBatchTimeout: config.ParseDurationEnvOrDefault("PRISME_EVENTSTORE_MAX_BATCH_TIMEOUT", 1*time.Minute),
	}
	service.pageViewRingBuf = ringo.NewWaiter(
		ringo.NewManyToOne(
			int(service.maxBatchSize*10),
			ringo.WithManyToOneCollisionHandler[*event.PageView](ringo.CollisionHandlerFunc(func(_ any) {
				service.logger.Warn().Msg("pageview events ring buffer collision detected, consider increasing PRISME_EVENTSTORE_MAX_BATCH_SIZE")
			})),
		),
	)
	service.customEventRingBuf = ringo.NewWaiter(
		ringo.NewManyToOne(
			int(service.maxBatchSize*10),
			ringo.WithManyToOneCollisionHandler[*event.Custom](ringo.CollisionHandlerFunc(func(_ any) {
				service.logger.Warn().Msg("custom events ring buffer collision detected, consider increasing PRISME_EVENTSTORE_MAX_BATCH_SIZE")
			})),
		),
	)

	logger = logger.With().
		Str("service", "eventstore").
		Str("service_impl", "clickhouse").
		Uint64("max_batch_size", service.maxBatchSize).
		Stringer("max_batch_timeout", service.maxBatchTimeout).
		Logger()

	logger.Info().Msg("clickhouse based event store configured")

	go service.batchPageViewLoop()
	go service.batchCustomEventLoop()

	return service
}

type ClickhouseService struct {
	logger             zerolog.Logger
	conn               driver.Conn
	maxBatchSize       uint64
	maxBatchTimeout    time.Duration
	pageViewRingBuf    ringo.Waiter[*event.PageView]
	customEventRingBuf ringo.Waiter[*event.Custom]
}

// StorePageView implements Service.
func (cs *ClickhouseService) StorePageView(_ context.Context, ev *event.PageView) error {
	cs.pageViewRingBuf.Push(ev)
	return nil
}

// StoreCustom implements Service.
func (cs *ClickhouseService) StoreCustom(_ context.Context, ev *event.Custom) error {
	cs.customEventRingBuf.Push(ev)
	return nil
}

func (cs *ClickhouseService) batchPageViewLoop() {
	var batch driver.Batch
	var err error
	batchCreationDate := time.Now()

	_ = batchCreationDate

	for {
		if batch == nil {
			batch, err = cs.conn.PrepareBatch(
				context.Background(),
				"INSERT INTO events_pageviews VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			)
			if err != nil {
				cs.logger.Err(err).Msg("failed to prepare next pageviews batch")
				continue
			}

			batchCreationDate = time.Now()
		}

		ev, done, dropped := cs.pageViewRingBuf.Next()
		if done {
			cs.logger.Info().Msg("page view ring buffer done, terminating batch loop...")
			return
		}
		if dropped > 0 {
			cs.logger.Info().Int("dropped", dropped).Msg("pageview events dropped")
		}

		cs.logger.Debug().Any("pageview_event", ev).Msg("appending pageview event to batch...")

		// Append to batch.
		err = batch.Append(
			ev.Timestamp,
			ev.PageUri.Host(),
			ev.PageUri.Path(),
			ev.Client.OperatingSystem,
			ev.Client.BrowserFamily,
			ev.Client.Device,
			ev.ReferrerUri.HostOrDirect(),
			ev.CountryCode,
		)
		if err != nil {
			cs.logger.Err(err).Msg("failed to append pageview to batch")
		}

		if uint64(batch.Rows()) >= cs.maxBatchSize || time.Since(batchCreationDate) > cs.maxBatchTimeout {
			go cs.sendBatch(batch)
			batch = nil
		}
	}
}

func (cs *ClickhouseService) batchCustomEventLoop() {
	var batch driver.Batch
	var err error
	batchCreationDate := time.Now()

	_ = batchCreationDate

	for {
		if batch == nil {
			batch, err = cs.conn.PrepareBatch(
				context.Background(),
				"INSERT INTO events_custom VALUES ($1, $2, $3, $4, $5, $6)",
			)
			if err != nil {
				cs.logger.Err(err).Msg("failed to prepare next custom events batch")
				continue
			}

			batchCreationDate = time.Now()
		}

		ev, done, dropped := cs.customEventRingBuf.Next()
		if done {
			cs.logger.Info().Msg("custom ring buffer done, terminating batch loop...")
			return
		}
		if dropped > 0 {
			cs.logger.Info().Int("dropped", dropped).Msg("custom events dropped")
		}

		cs.logger.Debug().Object("custom_event", ev).Msg("appending custom event to batch...")

		// Append to batch.
		err = batch.Append(
			ev.Timestamp,
			ev.PageUri.Host(),
			ev.PageUri.Path(),
			ev.Name,
			ev.Keys,
			ev.Values,
		)
		if err != nil {
			cs.logger.Err(err).Msg("failed to append custom event to batch")
		}

		if uint64(batch.Rows()) >= cs.maxBatchSize || time.Since(batchCreationDate) > cs.maxBatchTimeout {
			go cs.sendBatch(batch)
			batch = nil
		}
	}
}

func (cs *ClickhouseService) sendBatch(batch driver.Batch) {
	// Retry if an error occurred. This can happen on clickhouse cloud if instance
	// goes to idle state.
	var err error
	for i := 0; i < 5; i++ {
		err = batch.Send()
		if err != nil {
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			cs.logger.Debug().Msg("events batch successfully sent")
			break
		}
	}

	if err != nil {
		cs.logger.Err(err).Msg("failed to send events batch")
	}
}
