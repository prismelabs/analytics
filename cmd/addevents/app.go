package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
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

	for i := 0; i < a.cfg.BatchCount; i++ {
		batch, err := a.ch.Conn.PrepareBatch(ctx, "INSERT INTO prisme.events_pageviews")
		if err != nil {
			panic(err)
		}

		wg.Add(1)
		go func(i int, batch driver.Batch) {
			defer wg.Done()

			for j := 0; j < a.cfg.BatchSize; j++ {
				date := time.Now().Add(-randomMinute())
				date = date.Round(time.Minute)

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
		}(i, batch)
	}
	wg.Wait()
}

func (a App) AddCustomEvents() {
	ctx := context.Background()
	wg := sync.WaitGroup{}

	timeStep := time.Until(a.cfg.FromDate) / time.Duration(a.cfg.BatchCount)
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
				name, props := randomCustomEvent()

				err := batch.Append(
					date,
					randomItem(a.cfg.Domains),
					randomPathName(),
					name,
					props,
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

func randomCustomEvent() (string, string) {
	name := randomItem([]string{"click", "download", "sign_up", "subscription", "lot_of_props"})
	switch name {
	case "click":
		return name, fmt.Sprintf(`{"x":%v,"y":%v}`, rand.Intn(1024), rand.Intn(4096))
	case "download":
		return name, fmt.Sprintf(`{"doc":"%v.pdf"}`, randomString(alphaLower, 3))

	case "sign_up":
		return name, "{}"

	case "subscription":
		return name, fmt.Sprintf(`{"plan":"%v"}`, randomItem([]string{"growth", "premium", "enterprise"}))

	case "lot_of_props":
		propsCount := 128
		ev := strings.Builder{}
		ev.WriteRune('{')
		for i := 0; i < 128; i++ {
			ev.WriteString(`"`)
			ev.WriteString(randomString(alphaLower, 3))
			ev.WriteString(`"`)
			ev.WriteString(`:"`)
			ev.WriteString(randomString(alpha, 9))
			ev.WriteString(`"`)
			if i+1 < propsCount {
				ev.WriteString(`,`)
			}
		}
		ev.WriteRune('}')

		return name, ev.String()

	default:
		panic("not implemented")
	}
}
