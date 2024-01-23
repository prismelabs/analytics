package eventstore

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/prismelabs/prismeanalytics/internal/clickhouse"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/event"
	"github.com/prismelabs/prismeanalytics/internal/log"
)

// ClickhouseService define a clickhouse based event storage service.
type ClickhouseService struct {
	appendCh chan<- event.PageView
}

// ProvideClickhouseService is a wire provider for a clickhouse based event
// storage service.
func ProvideClickhouseService(ch clickhouse.Ch, logger log.Logger) Service {
	appendCh := make(chan event.PageView, 1024)

	batchSize := config.ParseUintEnvOrDefault("PRISME_EVENTSTORE_MAX_BATCH_SIZE", 4096, 64)
	batchTimeout := config.ParseDurationEnvOrDefault("PRISME_EVENTSTORE_MAX_BATCH_TIMEOUT", 1*time.Minute)

	logger.Info().
		Uint64("max_batch_size", batchSize).
		Dur("max_batch_timeout", batchTimeout).
		Msg("clickhouse based event store configured")

	go batchPageViewLoop(logger, ch.Conn, appendCh, batchSize, batchTimeout)

	return &ClickhouseService{
		appendCh,
	}
}

// StorePageViewEvent implements Service.
func (cs *ClickhouseService) StorePageViewEvent(ctx context.Context, ev event.PageView) error {
	cs.appendCh <- ev

	return nil
}

func batchPageViewLoop(logger log.Logger,
	conn driver.Conn,
	appendCh <-chan event.PageView,
	maxBatchSize uint64,
	maxBatchLifeTime time.Duration,
) {
	var batch driver.Batch
	var err error
	batchCreationDate := time.Now()

	for {
		if batch == nil {
			batch, err = conn.PrepareBatch(
				context.Background(),
				"INSERT INTO events_pageviews VALUES ($1, $2, $3, $4, $5, $6)",
			)
			if err != nil {
				logger.Err(err).Msg("failed to prepare batch")
			}

			batchCreationDate = time.Now()
		}

		ev := <-appendCh

		// Append to batch.
		err = batch.Append(
			ev.Timestamp,
			ev.DomainName,
			ev.PathName,
			ev.Client.OperatingSystem,
			ev.Client.BrowserFamily,
			ev.Client.Device,
		)
		if err != nil {
			logger.Err(err).Msg("failed to append to pageview batch")
		}

		if uint64(batch.Rows()) >= maxBatchSize || time.Since(batchCreationDate) > 3*time.Second {
			err := batch.Send()
			if err != nil {
				logger.Err(err).Msg("failed to send pageview batch")
			} else {
				logger.Debug().Msg("pageviews batch successfully sent")
			}
			batch = nil
		}
	}
}