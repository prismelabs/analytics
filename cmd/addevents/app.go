package main

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/log"
)

// NewApp returns a new App.
func NewApp(
	logger log.Logger,
	cfg Config,
	ch clickhouse.Ch,
) App {
	return App{
		logger: logger,
		cfg:    cfg,
		ch:     ch,
	}
}

// App contains application variables.
type App struct {
	logger log.Logger
	cfg    Config
	ch     clickhouse.Ch
}

func (a App) executeScenario(worker func(time.Time, Config, chan<- any) uint64) {
	timeStep := time.Since(a.cfg.FromDate) / time.Duration(a.cfg.BatchCount())
	totalEvents := &atomic.Uint64{}

	rowsChan := make(chan any)

	goPoolSize := max(runtime.NumCPU(), 4)
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
					a.logger.Fatal("failed to prepare sessions batch", err)
				}

				customEventsBatch, err := a.ch.PrepareBatch(context.Background(), "INSERT INTO prisme.events_custom")
				if err != nil {
					a.logger.Fatal("failed to prepare custom events batch", err)
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
						a.logger.Fatal("failed to append event to batch", err, reflect.TypeOf(record).Name())
					}
				}

				err = sessionsBatch.Send()
				if err != nil {
					a.logger.Fatal("failed to send sessions batch", err)
				}

				err = customEventsBatch.Send()
				if err != nil {
					a.logger.Fatal("failed to send custom events batch", err)
				}

				currentTotalEvent := totalEvents.Load()
				a.logger.Info(
					"batch done",
					"events", currentTotalEvent,
					"total_events", a.cfg.TotalEvents,
					"progress", fmt.Sprintf("%.2f%%", (float64(currentTotalEvent)/float64(a.cfg.TotalEvents))*100),
				)
			}

			wg.Done()
		}
	}

	wg.Wait()
}

func goPool(goroutines int) chan<- func() {
	ch := make(chan func(), goroutines)

	for range goroutines {
		go func() {
			for {
				task := <-ch
				task()
			}
		}()
	}

	return ch
}
