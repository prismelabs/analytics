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
					randomReferrerDomain(a.cfg.Domains),
					randomCountryCode(),
					fmt.Sprintf("prisme_%X", rand.Uint64()%uint64(a.cfg.BatchSize*a.cfg.BatchCount/3)),
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
					randomOS(),
					randomBrowser(),
					"benchbot",
					randomReferrerDomain(a.cfg.Domains),
					randomCountryCode(),
					fmt.Sprintf("prisme_%X", rand.Uint64()%uint64(a.cfg.BatchSize*a.cfg.BatchCount/3)),
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

var countryCodes = []string{
	"AF", "AX", "AL", "DZ", "AS", "AD", "AO", "AI", "AQ", "AG", "AR", "AM", "AW",
	"AU", "AT", "AZ", "BS", "BH", "BD", "BB", "BY", "BE", "BZ", "BJ", "BM", "BT",
	"BO", "BQ", "BA", "BW", "BV", "BR", "IO", "BN", "BG", "BF", "BI", "CV",
	"KH", "CM", "CA", "KY", "CF", "TD", "CL", "CN", "CX", "CC", "CO", "KM", "CD",
	"CG", "CK", "CR", "CI", "HR", "CU", "CW", "CY", "CZ", "DK", "DJ", "DM", "DO",
	"EC", "EG", "SV", "GQ", "ER", "EE", "SZ", "ET", "FK", "FO", "FJ", "FI", "FR",
	"GF", "PF", "TF", "GA", "GM", "GE", "DE", "GH", "GI", "GR", "GL", "GD", "GP",
	"GU", "GT", "GG", "GN", "GW", "GY", "HT", "HM", "VA", "HN", "HK", "HU", "IS",
	"IN", "ID", "IR", "IQ", "IE", "IM", "IL", "IT", "JM", "JP", "JE", "JO", "KZ",
	"KE", "KI", "KP", "KR", "KW", "KG", "LA", "LV", "LB", "LS", "LR", "LY", "LI",
	"LT", "LU", "MO", "MG", "MW", "MY", "MV", "ML", "MT", "MH", "MQ", "MR", "MU",
	"YT", "MX", "FM", "MD", "MC", "MN", "ME", "MS", "MA", "MZ", "MM", "NA", "NR",
	"NP", "NL", "NC", "NZ", "NI", "NE", "NG", "NU", "NF", "MK", "MP", "NO", "OM",
	"PK", "PW", "PS", "PA", "PG", "PY", "PE", "PH", "PN", "PL", "PT", "PR", "QA",
	"RE", "RO", "RU", "RW", "BL", "SH", "KN", "LC", "MF", "PM", "VC", "WS", "SM",
	"ST", "SA", "SN", "RS", "SC", "SL", "SG", "SX", "SK", "SI", "SB", "SO", "ZA",
	"GS", "SS", "ES", "LK", "SD", "SR", "SJ", "SE", "CH", "SY", "TW", "TJ", "TZ",
	"TH", "TL", "TG", "TK", "TO", "TT", "TN", "TR", "TM", "TC", "TV", "UG", "UA",
	"AE", "GB", "UM", "US", "UY", "UZ", "VU", "VE", "VN", "VG", "VI", "WF", "EH",
	"YE", "ZM", "ZW", "XX",
}

func randomCountryCode() string {
	return randomItem(countryCodes)
}
