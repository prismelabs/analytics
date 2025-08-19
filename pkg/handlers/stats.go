package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/handlers/utils"
	"github.com/prismelabs/analytics/pkg/services/stats"
)

type DataFrame[T any] struct {
	Keys   []T      `json:"keys"`
	Values []uint64 `json:"values"`
}

// Struct containing all /api/v1/stats/... handlers.
type Stats struct {
	Bounces          fiber.Handler
	Visitors         fiber.Handler
	Sessions         fiber.Handler
	SessionsDuration fiber.Handler
	PageViews        fiber.Handler
	LiveVisitors     fiber.Handler
	TopPages         fiber.Handler
	TopEntryPages    fiber.Handler
	TopExitPages     fiber.Handler
	TopReferrers     fiber.Handler
	TopUtmSources    fiber.Handler
	TopUtmMediums    fiber.Handler
	TopUtmCampaigns  fiber.Handler
	TopCountries     fiber.Handler
}

func GetStatsHandlers(s stats.Service) Stats {
	type TimeSerieFunc = func(
		stats.Service,
		context.Context,
		stats.Filters,
	) (stats.DataFrame[time.Time, uint64], error)

	newTimeSerieHandler := func(
		fetch TimeSerieFunc,
	) fiber.Handler {
		return func(c *fiber.Ctx) error {
			var err error

			filters, err := utils.ExtractStatsFilters(c)
			if err != nil {
				return err
			}

			df, err := fetch(s, c.UserContext(), filters)
			if err != nil {
				return err
			}

			return c.JSON(DataFrame[int64]{
				Keys:   timeToTimestamps(df.Keys),
				Values: df.Values,
			})
		}
	}

	type TopFunc = func(
		stats.Service,
		context.Context,
		stats.Filters,
		uint64,
	) (stats.DataFrame[string, uint64], error)
	newTopHandler := func(
		fetch TopFunc,
	) fiber.Handler {
		return func(c *fiber.Ctx) error {
			var err error

			filters, limit, err := utils.ExtractStatsFiltersAndLimit(c)
			if err != nil {
				return err
			}

			df, err := fetch(s, c.UserContext(), filters, limit)
			if err != nil {
				return err
			}

			return c.JSON(DataFrame[string](df))
		}
	}

	return Stats{
		Bounces:          newTimeSerieHandler(stats.Service.Bounces),
		Visitors:         newTimeSerieHandler(stats.Service.Visitors),
		Sessions:         newTimeSerieHandler(stats.Service.Sessions),
		SessionsDuration: newTimeSerieHandler(stats.Service.SessionsDuration),
		PageViews:        newTimeSerieHandler(stats.Service.PageViews),
		LiveVisitors:     newTimeSerieHandler(stats.Service.LiveVisitors),
		TopPages:         newTopHandler(stats.Service.TopPages),
		TopEntryPages:    newTopHandler(stats.Service.TopEntryPages),
		TopExitPages:     newTopHandler(stats.Service.TopExitPages),
		TopReferrers:     newTopHandler(stats.Service.TopReferrers),
		TopUtmSources:    newTopHandler(stats.Service.TopUtmSources),
		TopUtmMediums:    newTopHandler(stats.Service.TopUtmMediums),
		TopUtmCampaigns:  newTopHandler(stats.Service.TopUtmCampaigns),
		TopCountries:     newTopHandler(stats.Service.TopCountries),
	}
}

func timeToTimestamps(ti []time.Time) []int64 {
	ts := make([]int64, 0, cap(ti))
	for _, t := range ti {
		ts = append(ts, t.UnixMilli())
	}
	return ts
}
