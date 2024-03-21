package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/rs/zerolog"
)

func ProvideApp(logger zerolog.Logger, cfg Config, ch clickhouse.Ch) App {
	return App{
		logger: logger,
		cfg:    cfg,
		ch:     ch,
	}
}

type App struct {
	logger zerolog.Logger
	cfg    Config
	ch     clickhouse.Ch
}

func (a App) AddPageviewsEvents() {
	ctx := context.Background()
	wg := sync.WaitGroup{}

	timeStep := time.Since(a.cfg.FromDate) / time.Duration(a.cfg.BatchCount)
	timeCursor := a.cfg.FromDate

	for i := 0; i < a.cfg.BatchCount; i++ {
		batch, err := a.ch.Conn.PrepareBatch(ctx, "INSERT INTO prisme.events_pageviews")
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

				err := batch.Append(
					date,
					randomItem(a.cfg.Domains),
					randomPathName(),
					randomOS(),
					randomBrowser(),
					"benchbot",
					randomReferrerDomain(),
					"XX",
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

				err := batch.Append(
					date,
					randomItem(a.cfg.Domains),
					randomPathName(),
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

func randomCustomEvent() (string, []string, []string) {
	name := randomItem([]string{"click", "download", "sign_up", "subscription", "lot_of_props"})
	switch name {
	case "click":
		return name, []string{"x", "y"}, []string{fmt.Sprint(rand.Intn(3000)), fmt.Sprint(rand.Intn(2000))}
	case "download":
		return name, []string{"doc"}, []string{fmt.Sprintf("%v.pdf", randomString(alphaLower, 3))}

	case "sign_up":
		return name, []string{}, []string{}

	case "subscription":
		return name, []string{"plan"}, []string{randomItem([]string{"growth", "premium", "enterprise"})}

	case "lot_of_props":
		keys := make([]string, 64)
		values := make([]string, len(keys))
		for i := 0; i < len(keys); i++ {
			keys[i] = randomString(alphaLower, 3)
			values[i] = randomString(alpha, 9)
		}

		return name, keys, values

	default:
		panic("not implemented")
	}
}
