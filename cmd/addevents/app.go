package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/rs/zerolog"
)

// ProvideApp is a wire provider for App.
func ProvideApp(logger zerolog.Logger, cfg Config, ch clickhouse.Ch) App {
	return App{
		logger: logger,
		cfg:    cfg,
		ch:     ch,
	}
}

// App contains application variables.
type App struct {
	logger zerolog.Logger
	cfg    Config
	ch     clickhouse.Ch
}

func (a App) executeScenario(worker func(time.Time, Config, chan<- []any) uint64) {
	timeStep := time.Since(a.cfg.FromDate) / time.Duration(a.cfg.BatchCount())
	totalEvents := &atomic.Uint64{}

	rowsChan := make(chan []any)

	goPoolCh := goPool(runtime.NumCPU())

	// Create workers.
	go func() {
		for i := 0; i < runtime.NumCPU()/2; i++ {
			goPoolCh <- func() {
				timeCursor := a.cfg.FromDate
				step := uint64(0)

				for {
					for (totalEvents.Load() / uint64(a.cfg.BatchSize)) > step {
						timeCursor = timeCursor.Add(timeStep)
						step++
					}

					totalEvents.Add(worker(timeCursor, a.cfg, rowsChan))
				}
			}
		}
	}()

	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU()/2; i++ {
		wg.Add(1)
		goPoolCh <- func() {
			// Poll events and append them to the batch.
			for totalEvents.Load() < a.cfg.TotalEvents {
				batch, err := a.ch.PrepareBatch(context.Background(), "INSERT INTO prisme.sessions")
				if err != nil {
					a.logger.Panic().Err(err).Msg("failed to prepare sessions batch")
				}

				for i := uint64(0); i < a.cfg.BatchSize && totalEvents.Load() < a.cfg.TotalEvents; i++ {
					record := <-rowsChan
					if record == nil {
						break
					}

					err := batch.Append(record...)
					if err != nil {
						a.logger.Panic().Err(err).Msg("failed to append event to batch")
					}
				}

				err = batch.Send()
				if err != nil {
					a.logger.Panic().Err(err).Msg("failed to send batch")
				}

				currentTotalEvent := totalEvents.Load()
				a.logger.Info().
					Uint64("events", currentTotalEvent).
					Uint64("total_events", a.cfg.TotalEvents).
					Str("progress", fmt.Sprintf("%.2f%%", (float64(currentTotalEvent)/float64(a.cfg.TotalEvents))*100)).
					Msg("batch done")
			}

			wg.Done()
		}
	}

	wg.Wait()
}

// func (a App) pageviewBatch(ctx context.Context, batchId int, ch <-chan *Session) {
// 	batch, err := a.ch.Conn.PrepareBatch(ctx, "INSERT INTO prisme.sessions")
// 	if err != nil {
// 		a.logger.Panic().Err(err).Msg("failed to prepare batch")
// 	}
//
// 	for j := 0; j < a.cfg.BatchSize; j++ {
// 		session := <-ch
//
// 		err := batch.Append(
// 			session.domain,
// 			session.SessionTimestamp(),
// 			session.entryPath,
// 			session.exitTimestamp,
// 			session.exitPath,
// 			session.visitorId,
// 			session.sessionUuid,
// 			session.client.OperatingSystem,
// 			session.client.BrowserFamily,
// 			session.client.Device,
// 			session.referrerDomain,
// 			session.countryCode,
// 			session.utmSource,
// 			session.utmMedium,
// 			session.utmCampaign,
// 			session.utmTerm,
// 			session.utmContent,
// 			session.pageviews,
// 			session.sign,
// 		)
// 		if err != nil {
// 			a.logger.Panic().Err(err).Msg("failed to append to batch")
// 		}
// 	}
//
// 	err = batch.Send()
// 	if err != nil {
// 		a.logger.Panic().Err(err).Msg("failed to send to batch")
// 	}
// 	a.metrics.events.Add(uint64(a.cfg.BatchSize))
// 	a.logger.Info().
// 		Int("batch_count", a.cfg.BatchCount).
// 		Int("current_batch", batchId).
// 		Msg("batch done")
// }
//
// func (a App) AddCustomEvents() {
// 	ctx := context.Background()
// 	wg := sync.WaitGroup{}
//
// 	timeStep := time.Since(a.cfg.FromDate) / time.Duration(a.cfg.BatchCount)
// 	timeCursor := a.cfg.FromDate
//
// 	for i := 0; i < a.cfg.BatchCount; i++ {
// 		batch, err := a.ch.Conn.PrepareBatch(ctx, "INSERT INTO prisme.events_custom")
// 		if err != nil {
// 			panic(err)
// 		}
//
// 		// Move cursor.
// 		timeCursor = timeCursor.Add(timeStep)
//
// 		wg.Add(1)
// 		go func(i int, cursor time.Time, batch driver.Batch) {
// 			defer wg.Done()
//
// 			for j := 0; j < a.cfg.BatchSize; j++ {
// 				date := cursor.Add(-randomMinute())
// 				name, keys, values := randomCustomEvent()
//
// 				domain := randomItem(a.cfg.Domains)
//
// 				client := randomDesktopClient()
// 				err := batch.Append(
// 					date,
// 					domain,
// 					randomPathName(),
// 					client.OperatingSystem,
// 					client.BrowserFamily,
// 					client.Device,
// 					domain,
// 					randomCountryCode(),
// 					randomVisitorId(a.cfg.VisitorIdsRange),
// 					name,
// 					keys,
// 					values,
// 				)
// 				if err != nil {
// 					panic(err)
// 				}
// 			}
//
// 			err = batch.Send()
// 			if err != nil {
// 				panic(err)
// 			}
// 			a.logger.Info().
// 				Int("batch_count", a.cfg.BatchCount).
// 				Int("current_batch", i).
// 				Msg("batch done")
// 		}(i, timeCursor, batch)
// 	}
// 	wg.Wait()
// }

func goPool(goroutines int) chan<- func() {
	ch := make(chan func())

	for i := 0; i < goroutines; i++ {
		go func() {
			for {
				task := <-ch
				task()
			}
		}()
	}

	return ch
}
