package stats

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/prismelabs/analytics/pkg/services/eventdb"
	"github.com/prismelabs/analytics/pkg/services/teardown"
)

// Service define a statistics service.
type Service interface {
	Visitors(context.Context, Filters) (DataFrame[time.Time, uint64], error)
	Sessions(context.Context, Filters) (DataFrame[time.Time, uint64], error)
	SessionsDuration(context.Context, Filters) (DataFrame[time.Time, uint64], error)
	PageViews(context.Context, Filters) (DataFrame[time.Time, uint64], error)
	LiveVisitors(context.Context, Filters) (DataFrame[time.Time, uint64], error)
	Bounces(context.Context, Filters) (DataFrame[time.Time, uint64], error)
	TopPages(context.Context, Filters, uint64) (DataFrame[string, uint64], error)
	TopEntryPages(context.Context, Filters, uint64) (DataFrame[string, uint64], error)
	TopExitPages(context.Context, Filters, uint64) (DataFrame[string, uint64], error)
}

// DataFrame defines a columnar view over timestamped data.
type DataFrame[K, V any] struct {
	Keys   []K
	Values []V
}

// TimeRange define an interval of time.
type TimeRange struct {
	Start time.Time
	Dur   time.Duration
}

// Filters defines supported query filters.
type Filters struct {
	TimeRange       TimeRange
	Domain          []string
	Path            []string
	EntryPath       []string
	ExitPath        []string
	Referrers       []string
	OperatingSystem []string
	BrowserFamily   []string
	Country         []string
	UtmSource       []string
	UtmMedium       []string
	UtmCampaign     []string
	UtmTerm         []string
	UtmContent      []string
}

type service struct {
	db eventdb.Service
}

// NewService returns a new Service.
func NewService(
	db eventdb.Service,
	teardown teardown.Service,
) Service {
	return &service{db: db}
}

// Bounces implements Service.
func (s *service) Bounces(
	ctx context.Context,
	filters Filters,
) (DataFrame[time.Time, uint64], error) {
	var args []any
	var query = `
	WITH bounces AS (
		SELECT argMax(session_timestamp, version) AS session_timestamp,
		argMax(pageviews, version) AS pageviews
		FROM sessions
		WHERE session_id IN (` + sessionQuery(filters, &args) + `)
		GROUP BY session_id
		HAVING pageviews = 1
	)
	SELECT toStartOfInterval(toDateTime(session_timestamp), ` + s.interval(filters.TimeRange) + `) AS time,
		COUNT(*) AS bounces
	FROM bounces
	GROUP BY time
	ORDER BY time`

	return doQuery[time.Time](s.db, ctx, query, args...)
}

// LiveVisitors implements Service.
func (s *service) LiveVisitors(
	ctx context.Context,
	filters Filters,
) (DataFrame[time.Time, uint64], error) {
	var args []any
	var query = `
SELECT toStartOfInterval(toDateTime(session_timestamp), ` + s.interval(filters.TimeRange) + `) AS time,
	COUNT(DISTINCT(visitor_id))
FROM sessions
WHERE session_id IN (` + sessionQuery(filters, &args) + `)
AND addMinutes(exit_timestamp, 15) > ?
GROUP BY time
ORDER BY time`
	args = append(args, filters.TimeRange.Start.Add(filters.TimeRange.Dur))

	return doQuery[time.Time](s.db, ctx, query, args...)
}

// PageViews implements Service.
func (s *service) PageViews(
	ctx context.Context,
	filters Filters,
) (DataFrame[time.Time, uint64], error) {
	var args []any
	var query = `
SELECT toStartOfInterval(toDateTime(timestamp), ` + s.interval(filters.TimeRange) + `) AS time,
	COUNT(*)
FROM pageviews
WHERE session_id IN (` + sessionQuery(filters, &args) + `)
GROUP BY time
ORDER BY time`

	return doQuery[time.Time](s.db, ctx, query, args...)
}

// Sessions implements Service.
func (s *service) Sessions(
	ctx context.Context,
	filters Filters,
) (DataFrame[time.Time, uint64], error) {
	var args []any
	var query = `
SELECT toStartOfInterval(toDateTime(session_timestamp), ` + s.interval(filters.TimeRange) + `) AS time,
	COUNT(DISTINCT(session_id))
FROM sessions
WHERE session_id IN (` + sessionQuery(filters, &args) + `)
GROUP BY time
ORDER BY time`

	return doQuery[time.Time](s.db, ctx, query, args...)
}

// SessionsDuration implements Service.
func (s *service) SessionsDuration(
	ctx context.Context,
	filters Filters,
) (DataFrame[time.Time, uint64], error) {
	var args []any
	var query = `
WITH sessions_duration AS (
	SELECT toStartOfInterval(toDateTime(session_timestamp), ` + s.interval(filters.TimeRange) + `) AS time,
		argMax(session_timestamp, pageviews) as session_timestamp,
		argMax(exit_timestamp, pageviews) as exit_timestamp
	FROM sessions
	WHERE session_id IN (` + sessionQuery(filters, &args) + `)
	GROUP BY session_id
) SELECT time, toUInt64(avg(exit_timestamp - session_timestamp)) FROM sessions_duration
GROUP BY time
ORDER BY time`

	return doQuery[time.Time](s.db, ctx, query, args...)
}

// Visitors implements Service.
func (s *service) Visitors(
	ctx context.Context,
	filters Filters,
) (DataFrame[time.Time, uint64], error) {
	var args []any
	var query = `
SELECT toStartOfInterval(toDateTime(session_timestamp), ` + s.interval(filters.TimeRange) + `) AS time,
	COUNT(DISTINCT(visitor_id))
FROM sessions
WHERE session_id IN (` + sessionQuery(filters, &args) + `)
GROUP BY time
ORDER BY time`

	return doQuery[time.Time](s.db, ctx, query, args...)
}

// TopPages implements Service.
func (s *service) TopPages(ctx context.Context, filters Filters, limit uint64) (DataFrame[string, uint64], error) {
	var args []any
	var query = `
SELECT path, COUNT(*) AS pageviews
FROM pageviews
WHERE session_id IN (` + sessionQuery(filters, &args) + `)
AND ` + timeFilter("timestamp", filters, &args) + `
GROUP BY path
ORDER BY pageviews DESC
LIMIT ` + strconv.FormatUint(limit, 10)

	return doQuery[string](s.db, ctx, query, args...)
}

// TopEntryPages implements Service.
func (s *service) TopEntryPages(ctx context.Context, filters Filters, limit uint64) (DataFrame[string, uint64], error) {
	var args []any
	var query = `
WITH entry_pageviews AS (
	SELECT argMax(entry_path, pageviews) AS path
	FROM sessions
	WHERE session_id IN (` + sessionQuery(filters, &args) + `)
	AND ` + timeFilter("entry_timestamp", filters, &args) + `
	GROUP BY session_uuid
) SELECT path, COUNT(*) AS pageviews
FROM entry_pageviews
GROUP BY path
ORDER BY pageviews DESC
LIMIT ` + strconv.FormatUint(limit, 10)

	return doQuery[string](s.db, ctx, query, args...)
}

// TopExitPages implements Service.
func (s *service) TopExitPages(ctx context.Context, filters Filters, limit uint64) (DataFrame[string, uint64], error) {
	var args []any
	var query = `
WITH exit_pageviews AS (
	SELECT argMax(exit_path, pageviews) AS path
	FROM sessions
	WHERE session_id IN (` + sessionQuery(filters, &args) + `)
	AND ` + timeFilter("exit_timestamp", filters, &args) + `
	GROUP BY session_uuid
) SELECT path, COUNT(*) AS pageviews
FROM exit_pageviews
GROUP BY path
ORDER BY pageviews DESC
LIMIT ` + strconv.FormatUint(limit, 10)

	return doQuery[string](s.db, ctx, query, args...)
}

func (s *service) interval(timeRange TimeRange) string {
	if timeRange.Dur == 0 {
		return "INTERVAL 1 second"
	}

	return fmt.Sprintf("INTERVAL %d second", timeRange.Dur/(16*time.Second))
}

func doQuery[K any](
	db eventdb.Service,
	ctx context.Context,
	query string,
	args ...any,
) (DataFrame[K, uint64], error) {
	df := DataFrame[K, uint64]{
		Keys:   []K{},
		Values: []uint64{},
	}

	result, err := db.Query(ctx, query, args...)
	if err != nil {
		return DataFrame[K, uint64]{}, fmt.Errorf("query %v failed: %w", query, err)
	}

	for result.Next() {
		var k K
		var v uint64
		err := result.Scan(&k, &v)
		if err != nil {
			return DataFrame[K, uint64]{}, err
		}

		df.Keys = append(df.Keys, k)
		df.Values = append(df.Values, v)
	}

	return df, nil
}

func sessionQuery(filters Filters, args *[]any) string {
	query := "WITH filtered_sessions AS (SELECT *, session_id, pageviews FROM sessions WHERE 1 = 1"
	if (filters.TimeRange != TimeRange{}) {
		query += " AND (" + timeFilter("session_timestamp", filters, args)
		query += " OR " + timeFilter("exit_timestamp", filters, args) + ")"
	}

	query += ") SELECT session_id FROM filtered_sessions WHERE 1 = 1"

	if len(filters.Path) > 0 {
		query += " AND session_id IN (SELECT session_id FROM pageviews WHERE path IN ("
		for i, p := range filters.Path {
			if i > 0 {
				query += ", ?"
			} else {
				query += "?"
			}
			*args = append(*args, p)
		}
		query += ") AND session_id IN (select session_id FROM filtered_sessions))"
	}
	if len(filters.EntryPath) > 0 {
		query += " AND entry_path IN ("
		for i, p := range filters.EntryPath {
			if i > 0 {
				query += ", ?"
			} else {
				query += "?"
			}
			*args = append(*args, p)
		}
		query += ")"
	}
	if len(filters.ExitPath) > 0 {
		query += " AND exit_path IN ("
		for i, p := range filters.ExitPath {
			if i > 0 {
				query += ", ?"
			} else {
				query += "?"
			}
			*args = append(*args, p)
		}
		query += ")"
	}

	return query
}

func timeFilter(col string, filters Filters, args *[]any) string {
	timeRange := filters.TimeRange

	if (timeRange != TimeRange{}) {
		*args = append(
			*args,
			timeRange.Start.Format(time.DateTime),
			timeRange.Start.Add(timeRange.Dur).Format(time.DateTime),
		)

		return "(" + col + " >= toDateTime(?) AND " + col + " <= toDateTime(?))"
	}

	return "1 = 1"
}
