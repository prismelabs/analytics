package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/rs/zerolog"
)

// ProvideApp is a wire provider for App.
func ProvideApp(logger zerolog.Logger, cfg Config, source source.Driver, teardown teardown.Service) App {
	return App{
		logger: logger,
		cfg:    cfg,
		ch:     clickhouse.ProvideCh(logger, clickhouse.ProvideConfig(logger), source, teardown),
	}
}

// App contains application variables.
type App struct {
	logger zerolog.Logger
	cfg    Config
	ch     clickhouse.Ch
}

func (a App) executeScenario(worker func(time.Time, Config, chan<- any) uint64) {
	timeStep := time.Since(a.cfg.FromDate) / time.Duration(a.cfg.BatchCount())
	totalEvents := &atomic.Uint64{}

	rowsChan := make(chan any)

	goPoolSize := runtime.NumCPU()
	if goPoolSize < 4 {
		goPoolSize = 4
	}
	goPoolCh := goPool(goPoolSize)

	// Create workers.
	go func() {
		for i := 0; i < cap(goPoolCh)/2; i++ {
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

	for i := 0; i < cap(goPoolCh)/2; i++ {
		wg.Add(1)
		goPoolCh <- func() {
			// Poll events and append them to the batch.
			for totalEvents.Load() < a.cfg.TotalEvents {
				sessionsBatch, err := a.ch.PrepareBatch(context.Background(), "INSERT INTO prisme.sessions")
				if err != nil {
					a.logger.Panic().Err(err).Msg("failed to prepare sessions batch")
				}

				customEventsBatch, err := a.ch.PrepareBatch(context.Background(), "INSERT INTO prisme.events_custom")
				if err != nil {
					a.logger.Panic().Err(err).Msg("failed to prepare custom events batch")
				}

				for i := uint64(0); i < a.cfg.BatchSize && totalEvents.Load() < a.cfg.TotalEvents; i++ {
					record := <-rowsChan
					if record == nil {
						break
					}

					switch r := record.(type) {
					case Session:
						err = sessionsBatch.Append(r.Row()...)
					case CustomEvent:
						err = customEventsBatch.Append(r.Row()...)
					default:
						panic("unimplemented")
					}

					if err != nil {
						a.logger.Panic().Err(err).Type("record_type", record).Msg("failed to append event to batch")
					}
				}

				err = sessionsBatch.Send()
				if err != nil {
					a.logger.Panic().Err(err).Msg("failed to send sessions batch")
				}

				err = customEventsBatch.Send()
				if err != nil {
					a.logger.Panic().Err(err).Msg("failed to send custom events batch")
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

func goPool(goroutines int) chan<- func() {
	ch := make(chan func(), goroutines)

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
