//go:build test

package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/negrel/configue"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/event"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/eventdb"
	"github.com/prismelabs/analytics/pkg/services/eventstore"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
	"github.com/prismelabs/analytics/pkg/testutils/faker"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	logger := log.New("addevents", os.Stderr, true)
	err := logger.TestOutput()
	if err != nil {
		panic(err)
	}
	discardLogger := log.New("discard", io.Discard, false)

	figue := configue.New("", configue.ContinueOnError, configue.NewEnv("PRISME"), configue.NewFlag())
	var (
		config        Config
		clickhouseCfg clickhouse.Config
	)
	config.RegisterOptions(figue)
	clickhouseCfg.RegisterOptions(figue)

	err = figue.Parse()
	if err != nil {
		logger.Fatal("failed to parse configuration options", err)
	}
	err = clickhouseCfg.Validate()
	if err != nil {
		logger.Fatal("invalid options", err)
	}

	driver := clickhouse.EmbeddedSourceDriver(logger)
	teardown := teardown.NewService()
	db, err := eventdb.NewService(
		eventdb.Config{Driver: "clickhouse"},
		clickhouseCfg,
		logger,
		driver,
		teardown,
	)
	if err != nil {
		logger.Fatal("failed to create eventdb", err)
	}
	promRegistry := prometheus.NewRegistry()
	store, err := eventstore.NewService(eventstore.Config{
		MaxBatchSize:      10000,
		MaxBatchTimeout:   10 * time.Second,
		RingBuffersFactor: 10,
	}, db, discardLogger, promRegistry, teardown)
	if err != nil {
		logger.Fatal("failed to create eventstore", err)
	}

	start := time.Now()
	ctx := context.Background()
	exitRate := 1 / config.PageViewsPerSession
	var totalSessions atomic.Uint64

	// Report progress.
	go func() {
		for {
			time.Sleep(time.Second)
			total := totalSessions.Load()
			logger.Info("progress report",
				"actual_sessions", total,
				"total_sessions", config.TotalSessions,
				"progress", fmt.Sprintf("%.2f%%", float64(total)/float64(config.TotalSessions)*100),
			)
			if total >= config.TotalSessions {
				return
			}
		}
	}()

	var (
		wg sync.WaitGroup
		ch = make(chan func())
	)
	for range config.Workers {
		wg.Add(1)
		go worker(&wg, ch)
	}

	for range config.TotalSessions {
		ch <- func() {
			var (
				client uaparser.Client
				pv     event.PageView
			)

			if rand.Float64() < config.MobileRate {
				client = faker.UapMobileClient()
			} else {
				client = faker.UapDesktopClient()
			}

			session := event.Session{
				PageUri:       faker.Uri(),
				ReferrerUri:   faker.ReferrerUri(rand.Float64() < config.DirectTrafficRate),
				Client:        client,
				CountryCode:   faker.CountryCode(),
				VisitorId:     "prisme_" + faker.String(faker.AlphaNum, 16),
				SessionUuid:   faker.UuidV7(faker.Time(-6 * 31 * 24 * time.Hour)),
				Utm:           event.UtmParams{},
				PageviewCount: 0,
			}

			session.PageviewCount++
			pv = faker.PageView(session)
			_ = store.StorePageView(ctx, &pv)

			if rand.Float64() < config.BounceRate {
				goto endOfSession
			}

			for {
				session.PageviewCount++
				pv = faker.PageView(session)
				_ = store.StorePageView(ctx, &pv)

				if rand.Float64() < config.CustomEventsRate {
					switch rand.Intn(3) {
					case 0:
						ev := faker.CustomEvent(session)
						_ = store.StoreCustom(ctx, &ev)
					case 1:
						ev := faker.FileDownload(session)
						_ = store.StoreFileDownload(ctx, &ev)
					case 2:
						ev := faker.OutboundLinkClick(session)
						_ = store.StoreOutboundLinkClick(ctx, &ev)
					default:
						panic("oops")
					}
				}

				if rand.Float64() < exitRate {
					goto endOfSession
				}
			}

		endOfSession:
			totalSessions.Add(1)
		}
	}
	close(ch)
	wg.Wait()

	logger.Info("teardown...")
	err = teardown.Teardown()
	if err != nil {
		logger.Fatal("teardown failed", err)
	}

	logger.Info("scenario done", "duration", time.Since(start).String())
}

func worker(wg *sync.WaitGroup, ch <-chan func()) {
	for {
		work := <-ch
		if work == nil {
			wg.Done()
			return
		}

		work()
	}
}
