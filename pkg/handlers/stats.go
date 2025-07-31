package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/prismelabs/analytics/pkg/handlers/utils"
	"github.com/prismelabs/analytics/pkg/services/stats"
)

func GetStatsBatch(srv stats.Service) fiber.Handler {
	type DataFrame struct {
		Timestamps []int64  `json:"timestamps"`
		Values     []uint64 `json:"values"`
	}

	validMetrics := map[string]struct{}{
		"bounces":       {},
		"live-visitors": {},
		"pageviews":     {},
		"sessions":      {},
		"visitors":      {},
	}

	return func(c *fiber.Ctx) error {
		metrics := make(map[string]any)
		for _, m := range strings.Split(c.Query("metrics", ""), ",") {
			_, ok := validMetrics[m]
			if !ok {
				return fiber.NewError(fiber.StatusBadRequest, "unknown metrics")
			}

			metrics[m] = nil
		}

		timeRange, err := utils.ExtractTimeRange(c)
		if err != nil {
			return err
		}

		batch := srv.Begin(c.Context(), timeRange, stats.Filters{})
		defer batch.Close()

		// Compute metrics.
		for m, _ := range metrics {
			var df stats.DataFrame[uint64]
			switch m {
			case "bounces":
				df, err = batch.Bounces()
			case "live-visitors":
				df, err = batch.LiveVisitors()
			case "pageviews":
				df, err = batch.PageViews()
			case "sessions":
				df, err = batch.Sessions()
			case "visitors":
				df, err = batch.Visitors()
			}
			if err != nil {
				return fmt.Errorf("unexpected error occurred while computing metrics: %w", err)
			}

			ts := make([]int64, 0, cap(df.Timestamps))
			for _, t := range df.Timestamps {
				ts = append(ts, t.UnixMilli())
			}
			metrics[m] = DataFrame{
				Timestamps: ts,
				Values:     df.Values,
			}
		}

		return c.JSON(metrics)
	}
}
