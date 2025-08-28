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
	TopReferrers(context.Context, Filters, uint64) (DataFrame[string, uint64], error)
	TopUtmSources(context.Context, Filters, uint64) (DataFrame[string, uint64], error)
	TopUtmMediums(context.Context, Filters, uint64) (DataFrame[string, uint64], error)
	TopUtmCampaigns(context.Context, Filters, uint64) (DataFrame[string, uint64], error)
	TopCountries(context.Context, Filters, uint64) (DataFrame[string, uint64], error)
	TopBrowsers(context.Context, Filters, uint64) (DataFrame[string, uint64], error)
	TopOperatingSystems(context.Context, Filters, uint64) (DataFrame[string, uint64], error)
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

// TopReferrers implements Service.
func (s *service) TopReferrers(ctx context.Context, filters Filters, limit uint64) (DataFrame[string, uint64], error) {
	var args []any
	var query = `
WITH referrers AS (
	SELECT argMax(referrer_domain, pageviews) AS referrer
	FROM sessions
	WHERE session_id IN (` + sessionQuery(filters, &args) + `)
	GROUP BY session_uuid
) SELECT referrer, COUNT(*) AS sessions
FROM referrers
GROUP BY referrer
ORDER BY sessions DESC
LIMIT ` + strconv.FormatUint(limit, 10)

	return doQuery[string](s.db, ctx, query, args...)
}

// TopUtmSources implements Service.
func (s *service) TopUtmSources(ctx context.Context, filters Filters, limit uint64) (DataFrame[string, uint64], error) {
	var args []any
	var query = `
WITH utm_sources AS (
	SELECT argMax(utm_source, pageviews) AS utm_source
	FROM sessions
	WHERE session_id IN (` + sessionQuery(filters, &args) + `)
	GROUP BY session_uuid
) SELECT utm_source, COUNT(*) AS sessions
FROM utm_sources
GROUP BY utm_source
ORDER BY sessions DESC
LIMIT ` + strconv.FormatUint(limit, 10)

	return doQuery[string](s.db, ctx, query, args...)
}

// TopUtmMediums implements Service.
func (s *service) TopUtmMediums(ctx context.Context, filters Filters, limit uint64) (DataFrame[string, uint64], error) {
	var args []any
	var query = `
WITH utm_mediums AS (
	SELECT argMax(utm_medium, pageviews) AS utm_medium
	FROM sessions
	WHERE session_id IN (` + sessionQuery(filters, &args) + `)
	GROUP BY session_uuid
) SELECT utm_medium, COUNT(*) AS sessions
FROM utm_mediums
GROUP BY utm_medium
ORDER BY sessions DESC
LIMIT ` + strconv.FormatUint(limit, 10)

	return doQuery[string](s.db, ctx, query, args...)
}

// TopUtmCampaigns implements Service.
func (s *service) TopUtmCampaigns(ctx context.Context, filters Filters, limit uint64) (DataFrame[string, uint64], error) {
	var args []any
	var query = `
WITH utm_campaigns AS (
	SELECT argMax(utm_campaign, pageviews) AS utm_campaign
	FROM sessions
	WHERE session_id IN (` + sessionQuery(filters, &args) + `)
	GROUP BY session_uuid
) SELECT utm_campaign, COUNT(*) AS sessions
FROM utm_campaigns
GROUP BY utm_campaign
ORDER BY sessions DESC
LIMIT ` + strconv.FormatUint(limit, 10)

	return doQuery[string](s.db, ctx, query, args...)
}

// TopCountries implements Service.
func (s *service) TopCountries(ctx context.Context, filters Filters, limit uint64) (DataFrame[string, uint64], error) {
	var args []any
	var query = `
WITH sessions_locations AS (
	SELECT argMax(country_code, pageviews) AS country_code
	FROM sessions
	WHERE session_id IN (` + sessionQuery(filters, &args) + `)
	GROUP BY session_uuid
) SELECT country_code, COUNT(*) AS sessions
FROM sessions_locations
GROUP BY country_code
ORDER BY sessions DESC
LIMIT ` + strconv.FormatUint(limit, 10)

	return doQuery[string](s.db, ctx, query, args...)
}

// TopBrowsers implements Service.
func (s *service) TopBrowsers(ctx context.Context, filters Filters, limit uint64) (DataFrame[string, uint64], error) {
	var args []any
	var query = `
WITH sessions_browsers AS (
	SELECT argMax(browser_family, pageviews) AS browser
	FROM sessions
	WHERE session_id IN (` + sessionQuery(filters, &args) + `)
	GROUP BY session_uuid
) SELECT browser, COUNT(*) AS sessions
FROM sessions_browsers
GROUP BY browser
ORDER BY sessions DESC
LIMIT ` + strconv.FormatUint(limit, 10)

	return doQuery[string](s.db, ctx, query, args...)
}

// TopOperatingSystems implements Service.
func (s *service) TopOperatingSystems(ctx context.Context, filters Filters, limit uint64) (DataFrame[string, uint64], error) {
	var args []any
	var query = `
WITH sessions_os AS (
	SELECT argMax(operating_system, pageviews) AS os
	FROM sessions
	WHERE session_id IN (` + sessionQuery(filters, &args) + `)
	GROUP BY session_uuid
) SELECT os, COUNT(*) AS sessions
FROM sessions_os
GROUP BY os
ORDER BY sessions DESC
LIMIT ` + strconv.FormatUint(limit, 10)

	return doQuery[string](s.db, ctx, query, args...)
}

func (s *service) interval(timeRange TimeRange) string {
	if timeRange.Dur == 0 {
		return "INTERVAL 1 second"
	}

	return fmt.Sprintf("INTERVAL %d second", timeRange.Dur/(32*time.Second))
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
		query += " AND " + sessionTimeFilter(filters, args)
	}

	query += ") SELECT session_id FROM filtered_sessions WHERE 1 = 1"

	if len(filters.Domain) > 0 {
		query += " AND " + stringListFilter("domain", filters.Domain, args)
	}

	if len(filters.Path) > 0 {
		query += " AND session_id IN (SELECT session_id FROM pageviews WHERE " + stringListFilter("path", filters.Path, args)
		query += "AND session_id IN (select session_id FROM filtered_sessions))"
	}
	if len(filters.EntryPath) > 0 {
		query += " AND " + stringListFilter("entry_path", filters.EntryPath, args)
	}
	if len(filters.ExitPath) > 0 {
		query += " AND " + stringListFilter("exit_path", filters.ExitPath, args)
	}
	if len(filters.Referrers) > 0 {
		query += " AND " + stringListFilter("referrer_domain", filters.Referrers, args)
	}
	if len(filters.OperatingSystem) > 0 {
		query += " AND " + stringListFilter("operating_system", filters.OperatingSystem, args)
	}
	if len(filters.BrowserFamily) > 0 {
		query += " AND " + stringListFilter("browser_family", filters.BrowserFamily, args)
	}
	if len(filters.Country) > 0 {
		query += " AND " + stringListFilter("country_code", filters.Country, args)
	}
	if len(filters.UtmSource) > 0 {
		query += " AND " + stringListFilter("utm_source", filters.UtmSource, args)
	}
	if len(filters.UtmMedium) > 0 {
		query += " AND " + stringListFilter("utm_medium", filters.UtmMedium, args)
	}
	if len(filters.UtmCampaign) > 0 {
		query += " AND " + stringListFilter("utm_campaign", filters.UtmCampaign, args)
	}
	if len(filters.UtmTerm) > 0 {
		query += " AND " + stringListFilter("utm_term", filters.UtmTerm, args)
	}
	if len(filters.UtmContent) > 0 {
		query += " AND " + stringListFilter("utm_content", filters.UtmContent, args)
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

func sessionTimeFilter(filters Filters, args *[]any) string {
	timeRange := filters.TimeRange

	var query string
	query += "(" + timeFilter("session_timestamp", filters, args)
	query += " OR " + timeFilter("exit_timestamp", filters, args) + ")"
	query += " OR session_timestamp <= toDateTime(?) AND exit_timestamp >= toDateTime(?)"
	*args = append(*args,
		timeRange.Start.Format(time.DateTime),
		timeRange.Start.Add(timeRange.Dur).Format(time.DateTime),
	)

	return query
}

func stringListFilter(col string, list []string, args *[]any) string {
	query := col + " IN ("
	for i, p := range list {
		if i > 0 {
			query += ", ?"
		} else {
			query += "?"
		}
		*args = append(*args, p)
	}
	return query + ")"
}
