package eventstore

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/negrel/ringo"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// ProvideService is a wire provider for a clickhouse based event
// storage service.
func ProvideService(
	cfg Config,
	ch clickhouse.Ch,
	logger zerolog.Logger,
	promRegistry *prometheus.Registry,
	teardownService teardown.Service,
) Service {
	batchDone := make(chan struct{})
	logger = logger.With().
		Str("service", "eventstore").
		Str("service_impl", "clickhouse").
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
		<-batchDone
		<-batchDone
		<-batchDone
		logger.Info().Msg("event batch loops canceled.")
		return nil
	})

	service := &clickhouseService{
		logger:          logger,
		conn:            ch.Conn,
		maxBatchSize:    cfg.MaxBatchSize,
		maxBatchTimeout: cfg.MaxBatchTimeout,
		pageViewRingBuf: ringo.NewWaiter(
			ringo.NewManyToOne(
				int(cfg.MaxBatchSize*cfg.RingBuffersFactor),
				ringo.WithManyToOneCollisionHandler[*event.PageView](ringo.CollisionHandlerFunc(func(_ any) {
					logger.Warn().Msg("pageview events ring buffer collision detected, consider increasing PRISME_EVENTSTORE_RING_BUFFERS_FACTOR or PRISME_EVENTSTORE_MAX_BATCH_SIZE")
				})),
			),
			ringo.WithWaiterContext[*event.PageView](ctx),
		),
		customEventRingBuf: ringo.NewWaiter(
			ringo.NewManyToOne(
				int(cfg.MaxBatchSize*cfg.RingBuffersFactor),
				ringo.WithManyToOneCollisionHandler[*event.Custom](ringo.CollisionHandlerFunc(func(_ any) {
					logger.Warn().Msg("custom events ring buffer collision detected, consider increasing PRISME_EVENTSTORE_RING_BUFFERS_FACTOR or PRISME_EVENTSTORE_MAX_BATCH_SIZE")
				})),
			),
			ringo.WithWaiterContext[*event.Custom](ctx),
		),
		outboundLinkClickEventRingBuf: ringo.NewWaiter(
			ringo.NewManyToOne(
				int(cfg.MaxBatchSize*cfg.RingBuffersFactor),
				ringo.WithManyToOneCollisionHandler[*event.OutboundLinkClick](ringo.CollisionHandlerFunc(func(_ any) {
					logger.Warn().Msg("outbound link click events ring buffer collision detected, consider increasing PRISME_EVENTSTORE_RING_BUFFERS_FACTOR or PRISME_EVENTSTORE_MAX_BATCH_SIZE")
				})),
			),
			ringo.WithWaiterContext[*event.OutboundLinkClick](ctx),
		),
		fileDownloadEventRingBuf: ringo.NewWaiter(
			ringo.NewManyToOne(
				int(cfg.MaxBatchSize*cfg.RingBuffersFactor),
				ringo.WithManyToOneCollisionHandler[*event.FileDownload](ringo.CollisionHandlerFunc(func(_ any) {
					logger.Warn().Msg("file download events ring buffer collision detected, consider increasing PRISME_EVENTSTORE_RING_BUFFERS_FACTOR or PRISME_EVENTSTORE_MAX_BATCH_SIZE")
				})),
			),
			ringo.WithWaiterContext[*event.FileDownload](ctx),
		),
		metrics: newMetrics(promRegistry),
	}

	go service.batchPageViewLoop(batchDone)
	go service.batchCustomEventLoop(batchDone)
	go service.batchOutboundLinkClickEventLoop(batchDone)
	go service.batchFileDownloadEventLoop(batchDone)

	logger.Info().Msg("clickhouse based event store configured")

	return service
}

type clickhouseService struct {
	logger                        zerolog.Logger
	conn                          driver.Conn
	maxBatchSize                  uint64
	maxBatchTimeout               time.Duration
	pageViewRingBuf               ringo.Waiter[*event.PageView]
	customEventRingBuf            ringo.Waiter[*event.Custom]
	outboundLinkClickEventRingBuf ringo.Waiter[*event.OutboundLinkClick]
	fileDownloadEventRingBuf      ringo.Waiter[*event.FileDownload]
	metrics                       metrics
}

// StoreFileDownload implements Service.
func (cs *clickhouseService) StoreFileDownload(_ context.Context, ev *event.FileDownload) error {
	cs.fileDownloadEventRingBuf.Push(ev)
	return nil
}

// StoreOutboundLinkClick implements Service.
func (cs *clickhouseService) StoreOutboundLinkClick(_ context.Context, ev *event.OutboundLinkClick) error {
	cs.outboundLinkClickEventRingBuf.Push(ev)
	return nil
}

// StorePageView implements Service.
func (cs *clickhouseService) StorePageView(_ context.Context, ev *event.PageView) error {
	cs.pageViewRingBuf.Push(ev)
	return nil
}

// StoreCustom implements Service.
func (cs *clickhouseService) StoreCustom(_ context.Context, ev *event.Custom) error {
	cs.customEventRingBuf.Push(ev)
	return nil
}

func (cs *clickhouseService) batchPageViewLoop(batchDone chan<- struct{}) {
	var batch driver.Batch
	var err error
	batchCreationDate := time.Now()

	promLabels := prometheus.Labels{
		"type": "pageview",
	}

	for {
		if batch == nil {
			batch, err = cs.conn.PrepareBatch(
				context.Background(),
				// pageviews table is a materialized view derived from sessions.
				// sessions table engine is VersionedCollapsedMergeTree so we can
				// keep appending row with the same Session UUID.
				// See https://clickhouse.com/docs/en/engines/table-engines/mergetree-family/versionedcollapsingmergetree
				"INSERT INTO sessions",
			)
			if err != nil {
				cs.logger.Err(err).Msg("failed to prepare next pageviews batch")
				continue
			}

			batchCreationDate = time.Now()
		}

		// Wait for next event.
		ev, done, dropped := cs.pageViewRingBuf.Next()
		// Ring buffer context was canceled.
		if done {
			cs.logger.Info().Msg("page view events ring buffer done, sending last batch...")
			cs.sendBatch(batch, promLabels)
			cs.logger.Info().Msg("last batch of page view events sent.")
			batchDone <- struct{}{}
			return
		}
		if dropped > 0 {
			cs.logger.Info().Int("dropped", dropped).Msg("pageview events dropped")
			cs.metrics.droppedEvents.With(promLabels).Add(float64(dropped))
		}

		// Append to batch.
		cs.logger.Debug().Any("pageview_event", ev).Msg("appending pageview event to batch...")

		// Session already stored.
		if ev.Session.PageviewCount > 1 {
			// Cancel previous session.
			err = batch.Append(
				ev.Session.PageUri.Host(),
				ev.Session.PageUri.Path(),
				ev.Timestamp.UTC(),
				ev.PageUri.Path(),
				ev.Session.VisitorId,
				ev.Session.SessionUuid,
				ev.Session.Client.OperatingSystem,
				ev.Session.Client.BrowserFamily,
				ev.Session.Client.Device,
				ev.Session.ReferrerUri.HostOrDirect(),
				ev.Session.CountryCode,
				ev.Session.Utm.Source,
				ev.Session.Utm.Medium,
				ev.Session.Utm.Campaign,
				ev.Session.Utm.Term,
				ev.Session.Utm.Content,
				ev.Status,
				ev.Session.PageviewCount-1, // Cancel previous version.
				-1,
			)
			if err != nil {
				cs.logger.Err(err).Msg("failed to add cancel session row to batch")
			}
		}

		err = batch.Append(
			ev.Session.PageUri.Host(),
			ev.Session.PageUri.Path(),
			ev.Timestamp.UTC(),
			ev.PageUri.Path(),
			ev.Session.VisitorId,
			ev.Session.SessionUuid,
			ev.Session.Client.OperatingSystem,
			ev.Session.Client.BrowserFamily,
			ev.Session.Client.Device,
			ev.Session.ReferrerUri.HostOrDirect(),
			ev.Session.CountryCode,
			ev.Session.Utm.Source,
			ev.Session.Utm.Medium,
			ev.Session.Utm.Campaign,
			ev.Session.Utm.Term,
			ev.Session.Utm.Content,
			ev.Status,
			ev.Session.PageviewCount,
			1,
		)
		if err != nil {
			cs.logger.Err(err).Msg("failed to append pageview to batch")
		}

		if uint64(batch.Rows()) >= cs.maxBatchSize || time.Since(batchCreationDate) > cs.maxBatchTimeout {
			go cs.sendBatch(batch, promLabels)
			batch = nil
		}
	}
}

func (cs *clickhouseService) batchCustomEventLoop(batchDone chan<- struct{}) {
	var batch driver.Batch
	var err error
	batchCreationDate := time.Now()

	promLabels := prometheus.Labels{
		"type": "custom",
	}

	for {
		if batch == nil {
			batch, err = cs.conn.PrepareBatch(
				context.Background(),
				"INSERT INTO events_custom",
			)
			if err != nil {
				cs.logger.Err(err).Msg("failed to prepare next custom events batch")
				continue
			}

			batchCreationDate = time.Now()
		}

		// Wait for next event.
		ev, done, dropped := cs.customEventRingBuf.Next()
		// Ring buffer context was canceled.
		if done {
			cs.logger.Info().Msg("custom events ring buffer done, sending last batch...")
			cs.sendBatch(batch, promLabels)
			cs.logger.Info().Msg("last batch of custom events sent.")
			batchDone <- struct{}{}
			return
		}
		if dropped > 0 {
			cs.logger.Info().Int("dropped", dropped).Msg("custom events dropped")
			cs.metrics.droppedEvents.With(promLabels).Add(float64(dropped))
		}

		cs.logger.Debug().Object("custom_event", ev).Msg("appending custom event to batch...")

		// Append to batch.
		err = batch.Append(
			ev.Timestamp,
			ev.Session.PageUri.Host(),
			ev.PageUri.Path(),
			ev.Session.VisitorId,
			ev.Session.SessionUuid,
			ev.Name,
			ev.Keys,
			ev.Values,
		)
		if err != nil {
			cs.logger.Err(err).Msg("failed to append custom event to batch")
		}

		if uint64(batch.Rows()) >= cs.maxBatchSize || time.Since(batchCreationDate) > cs.maxBatchTimeout {
			go cs.sendBatch(batch, promLabels)
			batch = nil
		}
	}
}

func (cs *clickhouseService) batchOutboundLinkClickEventLoop(batchDone chan<- struct{}) {
	var batch driver.Batch
	var err error
	batchCreationDate := time.Now()

	promLabels := prometheus.Labels{
		"type": "outbound_link_click",
	}

	for {
		if batch == nil {
			batch, err = cs.conn.PrepareBatch(
				context.Background(),
				"INSERT INTO outbound_link_clicks",
			)
			if err != nil {
				cs.logger.Err(err).Msg("failed to prepare next outbound link click events batch")
				continue
			}

			batchCreationDate = time.Now()
		}

		// Wait for next event.
		ev, done, dropped := cs.outboundLinkClickEventRingBuf.Next()
		// Ring buffer context was canceled.
		if done {
			cs.logger.Info().Msg("outbound link click events ring buffer done, sending last batch...")
			cs.sendBatch(batch, promLabels)
			cs.logger.Info().Msg("last batch of outbound link click events sent.")
			batchDone <- struct{}{}
			return
		}
		if dropped > 0 {
			cs.logger.Info().Int("dropped", dropped).Msg("outbound link click events dropped")
			cs.metrics.droppedEvents.With(promLabels).Add(float64(dropped))
		}

		cs.logger.Debug().Object("outbound_link_click_event", ev).Msg("appending outbound link click event to batch...")

		// Append to batch.
		err = batch.Append(
			ev.Timestamp,
			ev.Session.PageUri.Host(),
			ev.PageUri.Path(),
			ev.Session.VisitorId,
			ev.Session.SessionUuid,
			ev.Link,
		)
		if err != nil {
			cs.logger.Err(err).Msg("failed to append outbound link click event to batch")
		}

		if uint64(batch.Rows()) >= cs.maxBatchSize || time.Since(batchCreationDate) > cs.maxBatchTimeout {
			go cs.sendBatch(batch, promLabels)
			batch = nil
		}
	}
}

func (cs *clickhouseService) batchFileDownloadEventLoop(batchDone chan<- struct{}) {
	var batch driver.Batch
	var err error
	batchCreationDate := time.Now()

	promLabels := prometheus.Labels{
		"type": "file_download",
	}

	for {
		if batch == nil {
			batch, err = cs.conn.PrepareBatch(
				context.Background(),
				"INSERT INTO file_downloads",
			)
			if err != nil {
				cs.logger.Err(err).Msg("failed to prepare next file download events batch")
				continue
			}

			batchCreationDate = time.Now()
		}

		// Wait for next event.
		ev, done, dropped := cs.fileDownloadEventRingBuf.Next()
		// Ring buffer context was canceled.
		if done {
			cs.logger.Info().Msg("file download events ring buffer done, sending last batch...")
			cs.sendBatch(batch, promLabels)
			cs.logger.Info().Msg("last batch of file download events sent.")
			batchDone <- struct{}{}
			return
		}
		if dropped > 0 {
			cs.logger.Info().Int("dropped", dropped).Msg("file download events dropped")
			cs.metrics.droppedEvents.With(promLabels).Add(float64(dropped))
		}

		cs.logger.Debug().Object("fild_download_event", ev).Msg("appending file download event to batch...")

		// Append to batch.
		err = batch.Append(
			ev.Timestamp,
			ev.Session.PageUri.Host(),
			ev.PageUri.Path(),
			ev.Session.VisitorId,
			ev.Session.SessionUuid,
			ev.FileUrl,
		)
		if err != nil {
			cs.logger.Err(err).Msg("failed to append file download event to batch")
		}

		if uint64(batch.Rows()) >= cs.maxBatchSize || time.Since(batchCreationDate) > cs.maxBatchTimeout {
			go cs.sendBatch(batch, promLabels)
			batch = nil
		}
	}
}

func (cs *clickhouseService) sendBatch(batch driver.Batch, labels prometheus.Labels) {
	// Retry if an error occurred. This can happen on clickhouse cloud if instance
	// goes to idle state.
	var err error
	for i := 0; i < 5; i++ {
		start := time.Now()

		err = batch.Send()
		if err != nil {
			time.Sleep(time.Duration(i) * time.Second)
			cs.metrics.batchRetry.With(labels).Inc()
		} else {
			cs.metrics.sendBatchDuration.With(labels).Observe(time.Since(start).Seconds())
			cs.metrics.batchSize.With(labels).Observe(float64(batch.Rows()))
			cs.metrics.eventsCounter.With(labels).Add(float64(batch.Rows()))
			cs.logger.Debug().Msg("events batch successfully sent")
			break
		}
	}

	if err != nil {
		cs.metrics.batchDropped.With(labels).Inc()
		cs.logger.Err(err).Msg("failed to send events batch")
	}
}
