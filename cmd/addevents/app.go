package main

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/gofiber/storage"
	"github.com/gofiber/storage/memory"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/rs/zerolog"
)

// ProvideApp is a wire provider for App.
func ProvideApp(logger zerolog.Logger, cfg Config, ch clickhouse.Ch) App {
	return App{
		logger:  logger,
		metrics: &Metrics{},
		cfg:     cfg,
		ch:      ch,
		storage: memory.New(memory.Config{
			GCInterval: time.Minute,
		}),
	}
}

// App contains application variables.
type App struct {
	logger  zerolog.Logger
	metrics *Metrics
	cfg     Config
	ch      clickhouse.Ch
	storage storage.Storage
}

func (a App) pageviewsScenario() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan any, a.cfg.BatchSize)

	timeStep := time.Since(a.cfg.FromDate) / time.Duration(a.cfg.BatchCount)

	// Create injector routines.
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			timeCursor := a.cfg.FromDate
			step := uint64(0)

			for {
				for (a.metrics.events.Load() / uint64(a.cfg.BatchSize)) > step {
					timeCursor = timeCursor.Add(timeStep)
					step++
				}

				a.pageviewVisitor(ch, timeCursor)
			}
		}()
	}

	var wg sync.WaitGroup

	goPoolCh := make(chan func(), runtime.NumCPU())
	// Start workers.
	for i := 0; i < cap(goPoolCh); i++ {
		go func() {
			for {
				fn := <-goPoolCh
				if fn == nil {
					break
				}
				fn()
			}
		}()
	}

	for i := 0; i < a.cfg.BatchCount; i++ {
		j := i
		goPoolCh <- func() {
			a.pageviewBatch(ctx, j, ch)
			wg.Done()
		}

		wg.Add(1)
	}

	wg.Wait()
}

func (a App) pageviewBatch(ctx context.Context, batchId int, ch <-chan any) {
	sessionBatch, err := a.ch.Conn.PrepareBatch(ctx, "INSERT INTO prisme.sessions")
	if err != nil {
		a.logger.Panic().Err(err).Msg("failed to prepare batch")
	}
	_ = sessionBatch

	pageViewBatch, err := a.ch.Conn.PrepareBatch(ctx, "INSERT INTO prisme.events_pageviews")
	if err != nil {
		a.logger.Panic().Err(err).Msg("failed to prepare batch")
	}

	for j := 0; j < a.cfg.BatchSize; j++ {
		ev := <-ch
		switch event := ev.(type) {
		case *Pageview:
			err := pageViewBatch.Append(
				event.timestamp,
				event.domain,
				event.pathname,
				event.visitorId,
				event.sessionId,
			)
			if err != nil {
				a.logger.Panic().Err(err).Msg("failed to append pageview to batch")
			}

		case *Session:
			err := sessionBatch.Append(
				event.timestamp,
				event.domain,
				event.pathname,
				event.os,
				event.browser,
				event.device,
				event.referrerDomain,
				event.countryCode,
				event.visitorId,
				event.sessionId,
			)
			if err != nil {
				a.logger.Panic().Err(err).Msg("failed to append session to batch")
			}

		default:
			panic(fmt.Errorf("unknown event type %v", reflect.TypeOf(ev)))
		}

	}

	err = sessionBatch.Send()
	if err != nil {
		a.logger.Panic().Err(err).Msg("failed to send session batch")
	}
	err = pageViewBatch.Send()
	if err != nil {
		a.logger.Panic().Err(err).Msg("failed to send pageview batch")
	}
	a.metrics.events.Add(uint64(a.cfg.BatchSize))
	a.logger.Info().
		Int("batch_count", a.cfg.BatchCount).
		Int("current_batch", batchId).
		Msg("batch done")
}

func (a App) AddCustomEvents() {
	ctx := context.Background()
	wg := sync.WaitGroup{}

	timeStep := time.Since(a.cfg.FromDate) / time.Duration(a.cfg.BatchCount)
	timeCursor := a.cfg.FromDate

	for i := 0; i < a.cfg.BatchCount; i++ {
		batch, err := a.ch.Conn.PrepareBatch(ctx, "INSERT INTO prisme.events_custom")
		if err != nil {
			panic(err)
		}

		// Move cursor.
		timeCursor = timeCursor.Add(timeStep)

		wg.Add(1)
		go func(i int, cursor time.Time, batch driver.Batch) {
			defer wg.Done()

			for j := 0; j < a.cfg.BatchSize; j++ {
				date := cursor.Add(-randomMinute())
				name, keys, values := randomCustomEvent()

				domain := randomItem(a.cfg.Domains)

				err := batch.Append(
					date,
					domain,
					randomPathName(),
					randomOS(),
					randomBrowser(),
					"benchbot",
					domain,
					randomCountryCode(),
					randomVisitorId(a.cfg.VisitorIdsRange),
					name,
					keys,
					values,
				)
				if err != nil {
					panic(err)
				}
			}

			err = batch.Send()
			if err != nil {
				panic(err)
			}
			a.logger.Info().
				Int("batch_count", a.cfg.BatchCount).
				Int("current_batch", i).
				Msg("batch done")
		}(i, timeCursor, batch)
	}
	wg.Wait()
}
