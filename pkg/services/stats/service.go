package stats

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/prismelabs/analytics/pkg/services/eventdb"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/sql"
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
	db        eventdb.Service
	tmpTables sync.Map
}

// NewService returns a new Service.
func NewService(
	db eventdb.Service,
	teardown teardown.Service,
) Service {
	srv := &service{db: db, tmpTables: sync.Map{}}
	teardown.RegisterProcedure(func() error {
		var errs []error

		srv.tmpTables.Range(func(key any, value any) bool {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			errs = append(errs, db.Exec(ctx, "DROP TABLE "+value.(string)))

			return true
		})

		return errors.Join(errs...)
	})

	return srv
}

// Bounces implements Service.
func (s *service) Bounces(
	ctx context.Context,
	filters Filters,
) (DataFrame[time.Time, uint64], error) {
	var b sql.Builder

	b.Str("WITH bounces AS (").
		Strs("SELECT argMax(session_timestamp, version) AS session_timestamp,",
			"argMax(pageviews, version) AS pageviews",
			"FROM sessions",
			"WHERE session_uuid IN (").Call(sessionQuery, filters).Strs(")",
		"GROUP BY session_uuid",
		"HAVING pageviews = 1",
		")").
		Str("SELECT toStartOfInterval(toDateTime(session_timestamp),").
		Call(interval, filters.TimeRange).
		Strs(") AS time,",
			"COUNT(*) as bounces",
			"FROM bounces",
			"GROUP BY time",
			"ORDER BY time")

	return doQuery[time.Time](s.db, ctx, &b)
}

// LiveVisitors implements Service.
func (s *service) LiveVisitors(
	ctx context.Context,
	filters Filters,
) (DataFrame[time.Time, uint64], error) {
	var b sql.Builder

	b.Str("SELECT toStartOfInterval(toDateTime(session_timestamp),").
		Call(interval, filters.TimeRange).Strs(") AS time,",
		"COUNT(DISTINCT(visitor_id))",
		"FROM sessions",
		"WHERE session_uuid IN (").Call(sessionQuery, filters).Str(")").
		Str("AND addMinutes(exit_timestamp, 15) > ?",
			filters.TimeRange.Start.Add(filters.TimeRange.Dur)).
		Strs("GROUP BY time",
			"ORDER BY time")

	return doQuery[time.Time](s.db, ctx, &b)
}

// PageViews implements Service.
func (s *service) PageViews(
	ctx context.Context,
	filters Filters,
) (DataFrame[time.Time, uint64], error) {
	var b sql.Builder

	b.Str("SELECT toStartOfInterval(toDateTime(timestamp),").
		Call(interval, filters.TimeRange).Str(") AS time,").
		Strs("COUNT(*)", "FROM pageviews", "WHERE session_uuid IN (").
		Call(sessionQuery, filters).Strs(")", "GROUP BY time", "ORDER BY time")

	return doQuery[time.Time](s.db, ctx, &b)
}

// Sessions implements Service.
func (s *service) Sessions(
	ctx context.Context,
	filters Filters,
) (DataFrame[time.Time, uint64], error) {
	var b sql.Builder

	b.Str("SELECT toStartOfInterval(toDateTime(session_timestamp),").
		Call(interval, filters.TimeRange).Str(") AS time,").
		Strs("COUNT(DISTINCT(session_uuid))",
			"FROM sessions",
			"WHERE session_uuid IN (").Call(sessionQuery, filters).Strs(")",
		"GROUP BY time",
		"ORDER BY time")

	return doQuery[time.Time](s.db, ctx, &b)
}

// SessionsDuration implements Service.
func (s *service) SessionsDuration(
	ctx context.Context,
	filters Filters,
) (DataFrame[time.Time, uint64], error) {
	var b sql.Builder

	b.Strs("WITH sessions_duration AS (",
		"  SELECT toStartOfInterval(toDateTime(session_timestamp),").
		Call(interval, filters.TimeRange).Str(") AS time,").
		Strs("argMax(session_timestamp, pageviews) as session_timestamp,",
			"argMax(exit_timestamp, pageviews) AS exit_timestamp",
			"FROM sessions",
			"WHERE session_uuid IN (").Call(sessionQuery, filters).Strs(")",
		"  GROUP BY session_uuid",
		")").
		Strs("SELECT time, toUInt64(avg(exit_timestamp - session_timestamp))",
			"FROM sessions_duration",
			"GROUP BY time",
			"ORDER BY time")

	return doQuery[time.Time](s.db, ctx, &b)
}

// Visitors implements Service.
func (s *service) Visitors(
	ctx context.Context,
	filters Filters,
) (DataFrame[time.Time, uint64], error) {
	var b sql.Builder

	b.Str("SELECT toStartOfInterval(toDateTime(session_timestamp),").
		Call(interval, filters.TimeRange).Str(") AS time,").
		Strs("COUNT(DISTINCT(visitor_id))",
			"FROM sessions",
			"WHERE session_uuid IN (").Call(sessionQuery, filters).Str(")").
		Strs("GROUP BY time",
			"ORDER BY time")

	return doQuery[time.Time](s.db, ctx, &b)
}

// TopPages implements Service.
func (s *service) TopPages(
	ctx context.Context,
	filters Filters,
	limit uint64,
) (DataFrame[string, uint64], error) {
	var b sql.Builder

	b.Strs("SELECT path, COUNT(*) AS pageviews FROM pageviews",
		"WHERE session_uuid IN (").Call(sessionQuery, filters).Str(")").
		Str("AND").Call(timeFilter, "timestamp", filters).
		Strs("GROUP BY path",
			"ORDER BY pageviews DESC",
		).Fmt("LIMIT %v", limit)

	return doQuery[string](s.db, ctx, &b)
}

// TopEntryPages implements Service.
func (s *service) TopEntryPages(
	ctx context.Context,
	filters Filters,
	limit uint64,
) (DataFrame[string, uint64], error) {
	var b sql.Builder

	b.Strs("WITH entry_pageviews AS (",
		"  SELECT argMax(entry_path, pageviews) AS path",
		"  FROM sessions",
		"  WHERE session_uuid IN (").Call(sessionQuery, filters).Str(")").
		Str("AND").Call(timeFilter, "entry_timestamp", filters).
		Strs("GROUP BY session_uuid",
			")").
		Strs(
			"SELECT path, COUNT(*) AS pageviews",
			"FROM entry_pageviews",
			"GROUP BY path",
			"ORDER BY pageviews DESC",
		).Fmt("LIMIT %v", limit)

	return doQuery[string](s.db, ctx, &b)
}

// TopExitPages implements Service.
func (s *service) TopExitPages(
	ctx context.Context,
	filters Filters,
	limit uint64,
) (DataFrame[string, uint64], error) {
	var b sql.Builder

	b.Strs(
		"WITH exit_pageviews AS (",
		"  SELECT argMax(exit_path, pageviews) AS path",
		"  FROM sessions",
		"  WHERE session_uuid IN (").Call(sessionQuery, filters).Str(")").
		Str("AND").Call(timeFilter, "exit_timestamp", filters).
		Strs(
			"GROUP BY session_uuid",
			")").
		Strs(
			"SELECT path, COUNT(*) AS pageviews",
			"FROM exit_pageviews",
			"GROUP BY path",
			"ORDER BY pageviews DESC",
		).Fmt("LIMIT %v", limit)

	return doQuery[string](s.db, ctx, &b)
}

// TopReferrers implements Service.
func (s *service) TopReferrers(
	ctx context.Context,
	filters Filters,
	limit uint64,
) (DataFrame[string, uint64], error) {
	var b sql.Builder

	b.Strs(
		"WITH referrers AS (",
		"  SELECT argMax(referrer_domain, pageviews) AS referrer",
		"  FROM sessions",
		"  WHERE session_uuid IN (").Call(sessionQuery, filters).Strs(")",
		"  GROUP BY session_uuid",
		")",
		"SELECT referrer, COUNT(*) AS sessions",
		"FROM referrers",
		"GROUP BY referrer",
		"ORDER BY sessions DESC",
	).Fmt("LIMIT %v", limit)

	return doQuery[string](s.db, ctx, &b)
}

// TopUtmSources implements Service.
func (s *service) TopUtmSources(
	ctx context.Context,
	filters Filters,
	limit uint64,
) (DataFrame[string, uint64], error) {
	var b sql.Builder

	b.Strs(
		"WITH utm_sources AS (",
		"  SELECT argMax(utm_source, pageviews) AS utm_source",
		"  FROM sessions",
		"  WHERE session_uuid IN (").Call(sessionQuery, filters).Str(")").
		Str("GROUP BY session_uuid").
		Str(")").
		Strs(
			"SELECT utm_source, COUNT(*) AS sessions",
			"FROM utm_sources",
			"GROUP BY utm_source",
			"ORDER BY sessions DESC",
		).Fmt("LIMIT %v", limit)

	return doQuery[string](s.db, ctx, &b)
}

// TopUtmMediums implements Service.
func (s *service) TopUtmMediums(
	ctx context.Context,
	filters Filters,
	limit uint64,
) (DataFrame[string, uint64], error) {
	var b sql.Builder

	b.Strs(
		"WITH utm_mediums AS (",
		"  SELECT argMax(utm_medium, pageviews) AS utm_medium",
		"  FROM sessions",
		"  WHERE session_uuid IN (").Call(sessionQuery, filters).Str(")").
		Strs(
			"GROUP BY session_uuid",
			")").
		Strs("SELECT utm_medium, COUNT(*) AS sessions",
			"FROM utm_mediums",
			"GROUP BY utm_medium",
			"ORDER BY sessions DESC",
		).Fmt("LIMIT %v", limit)

	return doQuery[string](s.db, ctx, &b)
}

// TopUtmCampaigns implements Service.
func (s *service) TopUtmCampaigns(
	ctx context.Context,
	filters Filters,
	limit uint64,
) (DataFrame[string, uint64], error) {
	var b sql.Builder

	b.Strs(
		"WITH utm_campaigns AS (",
		"  SELECT argMax(utm_campaign, pageviews) AS utm_campaign",
		"  FROM sessions",
		"  WHERE session_uuid IN (").Call(sessionQuery, filters).Str(")").
		Strs(
			"GROUP BY session_uuid",
			")",
		).Strs(
		"SELECT utm_campaign, COUNT(*) AS sessions",
		"FROM utm_campaigns",
		"GROUP BY utm_campaign",
		"ORDER BY sessions DESC",
	).Fmt("LIMIT %v", limit)

	return doQuery[string](s.db, ctx, &b)
}

// TopCountries implements Service.
func (s *service) TopCountries(
	ctx context.Context,
	filters Filters,
	limit uint64,
) (DataFrame[string, uint64], error) {
	var b sql.Builder

	b.Strs(
		"WITH sessions_locations AS (",
		"  SELECT argMax(country_code, pageviews) AS country_code",
		"  FROM sessions",
		"  WHERE session_uuid IN (").Call(sessionQuery, filters).Str(")").
		Strs("GROUP BY session_uuid",
			")").
		Strs(
			"SELECT country_code, COUNT(*) AS sessions",
			"FROM sessions_locations",
			"GROUP BY country_code",
			"ORDER BY sessions DESC",
		).Fmt("LIMIT %v", limit)

	return doQuery[string](s.db, ctx, &b)
}

// TopBrowsers implements Service.
func (s *service) TopBrowsers(
	ctx context.Context,
	filters Filters,
	limit uint64,
) (DataFrame[string, uint64], error) {
	var b sql.Builder

	b.Strs(
		"WITH sessions_browsers AS (",
		"  SELECT argMax(browser_family, pageviews) AS browser",
		"  FROM sessions",
		"  WHERE session_uuid IN (").Call(sessionQuery, filters).Str(")").
		Strs("GROUP BY session_uuid",
			")").
		Strs(
			"SELECT browser, COUNT(*) AS sessions",
			"FROM sessions_browsers",
			"GROUP BY browser",
			"ORDER BY sessions DESC",
		).Fmt("LIMIT %v", limit)

	return doQuery[string](s.db, ctx, &b)
}

// TopOperatingSystems implements Service.
func (s *service) TopOperatingSystems(
	ctx context.Context,
	filters Filters,
	limit uint64,
) (DataFrame[string, uint64], error) {
	var b sql.Builder

	b.Strs(
		"WITH sessions_os AS (",
		"  SELECT argMax(operating_system, pageviews) AS os",
		"  FROM sessions",
		"  WHERE session_uuid IN (").Call(sessionQuery, filters).Str(")").
		Strs("GROUP BY session_uuid",
			")").
		Strs(
			"SELECT os, COUNT(*) AS sessions",
			"FROM sessions_os",
			"GROUP BY os",
			"ORDER BY sessions DESC",
		).Fmt("LIMIT %v", limit)

	return doQuery[string](s.db, ctx, &b)
}

func interval(builder *sql.Builder, args ...any) {
	timeRange := args[0].(TimeRange)

	if timeRange.Dur == 0 {
		builder.Str("INTERVAL 1 second")
	} else {
		builder.Str(
			fmt.Sprintf("INTERVAL %d second", timeRange.Dur/(32*time.Second)),
		)
	}
}

func doQuery[K any](
	db eventdb.Service,
	ctx context.Context,
	builder *sql.Builder,
) (DataFrame[K, uint64], error) {
	df := DataFrame[K, uint64]{
		Keys:   []K{},
		Values: []uint64{},
	}

	query, args := builder.Finish()

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

func sessionQuery(builder *sql.Builder, args ...any) {
	var (
		sub     sql.Builder
		filters = args[0].(Filters)
	)

	sub.Str("SELECT session_uuid FROM sessions WHERE 1 = 1")
	if (filters.TimeRange != TimeRange{}) {
		sub.Str("AND (").Call(sessionTimeFilter, filters).Str(")")
	}
	if len(filters.Domain) > 0 {
		sub.Str("AND").Call(stringListFilter, "domain", filters.Domain)
	}
	if len(filters.EntryPath) > 0 {
		sub.Str("AND").Call(stringListFilter, "entry_path", filters.EntryPath)
	}
	if len(filters.ExitPath) > 0 {
		sub.Str("AND").Call(stringListFilter, "exit_path", filters.ExitPath)
	}
	if len(filters.Referrers) > 0 {
		sub.Str("AND").Call(stringListFilter, "referrer_domain", filters.Referrers)
	}
	if len(filters.OperatingSystem) > 0 {
		sub.Str("AND").Call(stringListFilter, "operating_system", filters.OperatingSystem)
	}
	if len(filters.BrowserFamily) > 0 {
		sub.Str("AND").Call(stringListFilter, "browser_family", filters.BrowserFamily)
	}
	if len(filters.Country) > 0 {
		sub.Str("AND").Call(stringListFilter, "country_code", filters.Country)
	}
	if len(filters.UtmSource) > 0 {
		sub.Str("AND").Call(stringListFilter, "utm_source", filters.UtmSource)
	}
	if len(filters.UtmMedium) > 0 {
		sub.Str("AND").Call(stringListFilter, "utm_medium", filters.UtmMedium)
	}
	if len(filters.UtmCampaign) > 0 {
		sub.Str("AND").Call(stringListFilter, "utm_campaign", filters.UtmCampaign)
	}
	if len(filters.UtmTerm) > 0 {
		sub.Str("AND").Call(stringListFilter, "utm_term", filters.UtmTerm)
	}
	if len(filters.UtmContent) > 0 {
		sub.Str("AND").Call(stringListFilter, "utm_content", filters.UtmContent)
	}

	if len(filters.Path) > 0 {
		query, args := sub.Finish()

		sub.Str("AND session_uuid IN (SELECT session_uuid FROM pageviews WHERE").
			Call(stringListFilter, "path", filters.Path).
			Str("AND session_uuid IN (").Str(query, args...).Str("))")
	}

	query, args := sub.Finish()
	builder.Str(query, args...)
}

func timeFilter(builder *sql.Builder, args ...any) {
	col := args[0].(string)
	filters := args[1].(Filters)

	timeRange := filters.TimeRange

	if (timeRange != TimeRange{}) {
		builder.
			Strs("(", col, " >= ").
			Str("toDateTime(?)", timeRange.Start.Format(time.DateTime)).
			Strs("AND", col, " <= ").
			Str("toDateTime(?))", timeRange.Start.Add(timeRange.Dur).Format(time.DateTime))
	}
}

func sessionTimeFilter(builder *sql.Builder, args ...any) {
	filters := args[0].(Filters)

	timeRange := filters.TimeRange

	builder.Str("(").
		Call(timeFilter, "session_timestamp", filters).
		Str("OR").
		Call(timeFilter, "exit_timestamp", filters).
		Str(")").Str(
		"OR (session_timestamp <= toDateTime(?) AND exit_timestamp >= toDateTime(?))",
		timeRange.Start.Format(time.DateTime),
		timeRange.Start.Add(timeRange.Dur).Format(time.DateTime),
	)
}

func stringListFilter(builder *sql.Builder, args ...any) {
	col := args[0].(string)
	list := args[1].([]string)

	builder.Strs(col, "IN (")
	for i, p := range list {
		if i > 0 {
			builder.Str(", ?", p)
		} else {
			builder.Str("?", p)
		}
	}
	builder.Str(")")
}
