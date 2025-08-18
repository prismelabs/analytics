package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/handlers/utils"
	"github.com/prismelabs/analytics/pkg/services/stats"
)

type DataFrame[T any] struct {
	Keys   []T      `json:"keys"`
	Values []uint64 `json:"values"`
}

// GetStatsBounces returns a GET /api/v1/stats/bounces handler.
func GetStatsBounces(s stats.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var err error

		filters, err := utils.ExtractStatsFilters(c)
		if err != nil {
			return err
		}

		df, err := s.Bounces(c.UserContext(), filters)
		if err != nil {
			return err
		}

		return c.JSON(DataFrame[int64]{
			Keys:   timeToTimestamps(df.Keys),
			Values: df.Values,
		})
	}
}

// GetStatsVisitors returns a GET /api/v1/stats/visitors handler.
func GetStatsVisitors(s stats.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var err error

		filters, err := utils.ExtractStatsFilters(c)
		if err != nil {
			return err
		}

		df, err := s.Visitors(c.UserContext(), filters)
		if err != nil {
			return err
		}

		return c.JSON(DataFrame[int64]{
			Keys:   timeToTimestamps(df.Keys),
			Values: df.Values,
		})
	}
}

// GetStatsSessions returns a GET /api/v1/stats/sessions handler.
func GetStatsSessions(s stats.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var err error

		filters, err := utils.ExtractStatsFilters(c)
		if err != nil {
			return err
		}

		df, err := s.Sessions(c.UserContext(), filters)
		if err != nil {
			return err
		}

		return c.JSON(DataFrame[int64]{
			Keys:   timeToTimestamps(df.Keys),
			Values: df.Values,
		})
	}
}

// GetStatsSessionsDuration returns a GET /api/v1/stats/sessions-duration handler.
func GetStatsSessionsDuration(s stats.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var err error

		filters, err := utils.ExtractStatsFilters(c)
		if err != nil {
			return err
		}

		df, err := s.SessionsDuration(c.UserContext(), filters)
		if err != nil {
			return err
		}

		return c.JSON(DataFrame[int64]{
			Keys:   timeToTimestamps(df.Keys),
			Values: df.Values,
		})
	}
}

// GetStatsPageViews returns a GET /api/v1/stats/pageviews handler.
func GetStatsPageViews(s stats.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var err error

		filters, err := utils.ExtractStatsFilters(c)
		if err != nil {
			return err
		}

		df, err := s.PageViews(c.UserContext(), filters)
		if err != nil {
			return err
		}

		return c.JSON(DataFrame[int64]{
			Keys:   timeToTimestamps(df.Keys),
			Values: df.Values,
		})
	}
}

// GetStatsLiveVisitors returns a GET /api/v1/stats/live-visitors handler.
func GetStatsLiveVisitors(s stats.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var err error

		filters, err := utils.ExtractStatsFilters(c)
		if err != nil {
			return err
		}

		df, err := s.LiveVisitors(c.UserContext(), filters)
		if err != nil {
			return err
		}

		return c.JSON(DataFrame[int64]{
			Keys:   timeToTimestamps(df.Keys),
			Values: df.Values,
		})
	}
}

// GetStatsTopPages returns a GET /api/v1/stats/top-pages handler.
func GetStatsTopPages(s stats.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := strconv.ParseUint(c.Query("limit", "10"), 10, 64)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		filters, err := utils.ExtractStatsFilters(c)
		if err != nil {
			return err
		}

		df, err := s.TopPages(c.UserContext(), filters, limit)
		if err != nil {
			return err
		}

		return c.JSON(DataFrame[string](df))
	}
}

// GetStatsTopEntryPages returns a GET /api/v1/stats/top-entry-pages handler.
func GetStatsTopEntryPages(s stats.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := strconv.ParseUint(c.Query("limit", "10"), 10, 64)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		filters, err := utils.ExtractStatsFilters(c)
		if err != nil {
			return err
		}

		df, err := s.TopEntryPages(c.UserContext(), filters, limit)
		if err != nil {
			return err
		}

		return c.JSON(DataFrame[string](df))
	}
}

// GetStatsTopExitPages returns a GET /api/v1/stats/top-exit-pages handler.
func GetStatsTopExitPages(s stats.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, err := strconv.ParseUint(c.Query("limit", "10"), 10, 64)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		filters, err := utils.ExtractStatsFilters(c)
		if err != nil {
			return err
		}

		df, err := s.TopExitPages(c.UserContext(), filters, limit)
		if err != nil {
			return err
		}

		return c.JSON(DataFrame[string](df))
	}
}

func timeToTimestamps(ti []time.Time) []int64 {
	ts := make([]int64, 0, cap(ti))
	for _, t := range ti {
		ts = append(ts, t.UnixMilli())
	}
	return ts
}
