package stats

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/prismelabs/analytics/pkg/services/eventdb"
)

// Filters define statistics options.
type Filters struct {
	Domain          []string
	Path            []string
	EntryPath       []string
	ExitPath        []string
	Referrals       []string
	OperatingSystem []string
	BrowserFamily   []string
	Country         []string
	UtmSource       []string
	UtmMedium       []string
	UtmCampaign     []string
	UtmTerm         []string
	UtmContent      []string
}

// TimeRange define an interval of time.
type TimeRange struct {
	Start time.Time
	Dur   time.Duration
}

// Service define a statistics service.
type Service interface {
	Begin(context.Context, TimeRange, Filters) Batch
}

// DataFrame defines a columnar view over timestamped data.
type DataFrame[T any] struct {
	Timestamps []time.Time
	Values     []T
}

type Batch interface {
	Visitors() (DataFrame[uint64], error)
	Sessions() (DataFrame[uint64], error)
	PageViews() (DataFrame[uint64], error)
	LiveVisitors() (DataFrame[uint64], error)
	Bounces() (DataFrame[uint64], error)
	Close() error
}

type service struct {
	db eventdb.Service
}

type batch struct {
	tablePrefix string
	ctx         context.Context
	db          eventdb.Service
	timeRange   TimeRange
	filters     Filters
}

// NewService returns a new Service.
func NewService(db eventdb.Service) Service {
	return &service{db}
}

// Begin implements Service.
func (s *service) Begin(ctx context.Context, timeRange TimeRange, filters Filters) Batch {
	return &batch{"", ctx, s.db, timeRange, filters}
}

// Bounces implements Batch.
func (b *batch) Bounces() (DataFrame[uint64], error) {
	tab, err := b.sessionTable()
	if err != nil {
		return DataFrame[uint64]{}, err
	}

	var query = `
SELECT toStartOfInterval(toDateTime(session_timestamp), ` + b.interval() + `) AS time,
	COUNT(*)
FROM ` + tab + `
WHERE pageviews = 1
GROUP BY time
ORDER BY time`

	return b.query(query)
}

// Close implements Batch.
func (b *batch) Close() error {
	return b.db.Exec(b.ctx, "DROP TABLE "+b.tmpTableName("sessions"))
}

// LiveVisitors implements Batch.
func (b *batch) LiveVisitors() (DataFrame[uint64], error) {
	tab, err := b.sessionTable()
	if err != nil {
		return DataFrame[uint64]{}, err
	}

	var query = `
SELECT toStartOfInterval(toDateTime(session_timestamp), ` + b.interval() + `) AS time,
	COUNT(DISTINCT(visitor_id))
FROM ` + tab + `
WHERE addMinutes(exit_timestamp, 15) > now()
GROUP BY time
ORDER BY time`

	return b.query(query)
}

// PageViews implements Batch.
func (b *batch) PageViews() (DataFrame[uint64], error) {
	tab, err := b.sessionTable()
	if err != nil {
		return DataFrame[uint64]{}, err
	}

	var query = `
SELECT toStartOfInterval(toDateTime(timestamp), ` + b.interval() + `) AS time,
	COUNT(*)
FROM pageviews
WHERE session_id IN (
	SELECT session_id FROM ` + tab + `
)
GROUP BY time
ORDER BY time`

	return b.query(query)
}

// Sessions implements Batch.
func (b *batch) Sessions() (DataFrame[uint64], error) {
	tab, err := b.sessionTable()
	if err != nil {
		return DataFrame[uint64]{}, err
	}

	var query = `
SELECT toStartOfInterval(toDateTime(session_timestamp), ` + b.interval() + `) AS time,
	COUNT(*)
FROM ` + tab + `
GROUP BY time
ORDER BY time`

	return b.query(query)
}

// Visitors implements Batch.
func (b *batch) Visitors() (DataFrame[uint64], error) {
	tab, err := b.sessionTable()
	if err != nil {
		return DataFrame[uint64]{}, err
	}

	var query = `
SELECT toStartOfInterval(toDateTime(session_timestamp), ` + b.interval() + `) AS time,
	COUNT(DISTINCT(visitor_id))
FROM ` + tab + `
GROUP BY time
ORDER BY time`

	return b.query(query)
}

func (b *batch) tmpTableName(name string) string {
	if b.tablePrefix == "" {
		b.tablePrefix = "tmp_stats_" + strconv.FormatUint(rand.Uint64(), 16)
	}

	return b.tablePrefix + "_" + name
}

// sessionTable is an idempotent method that creates and initializes temporary
// session table.
func (b *batch) sessionTable() (string, error) {
	tabName := b.tmpTableName("sessions")

	row := b.db.QueryRow(b.ctx, "SELECT 1 FROM "+tabName)
	err := row.Err()
	if err != nil {
		if strings.Contains(err.Error(), "Unknown table expression identifier 'tmp_stats_") {
			// Table doesn't exists, create it.
			err = b.db.Exec(b.ctx, "CREATE TABLE "+tabName+" AS sessions ENGINE = Memory;")
			if err != nil {
				return "", fmt.Errorf("failed to create temporary session table: %w", err)
			}

			// Build query.
			query := "INSERT INTO " + tabName + " SELECT * FROM sessions WHERE 1 = 1"
			var args []any
			if (b.timeRange != TimeRange{}) {
				query += " AND (session_timestamp >= toDateTime(?) and session_timestamp <= toDateTime(?) OR exit_timestamp >= toDateTime(?) AND exit_timestamp <= toDateTime(?))"
				args = append(args, b.timeRange.Start, b.timeRange.Start.Add(b.timeRange.Dur), b.timeRange.Start, b.timeRange.Start.Add(b.timeRange.Dur))
			}

			// Insert into table.
			err := b.db.Exec(b.ctx, query, args...)
			if err != nil {
				return "", fmt.Errorf("failed to insert sessions into stats temporary table: %w", err)
			}
		} else {
			return "", fmt.Errorf("failed to check existence of stats temporary table: %w", err)
		}
	}

	return tabName, nil
}

func (b *batch) interval() string {
	if b.timeRange.Dur == 0 {
		return "INTERVAL 1 second"
	}

	return fmt.Sprintf("INTERVAL %d second", b.timeRange.Dur/(16*time.Second))
}

func (b *batch) query(query string, args ...any) (DataFrame[uint64], error) {
	df := DataFrame[uint64]{
		Timestamps: []time.Time{},
		Values:     []uint64{},
	}

	result, err := b.db.Query(b.ctx, query, args...)
	if err != nil {
		return DataFrame[uint64]{}, err
	}

	for result.Next() {
		var ti time.Time
		var v uint64
		err := result.Scan(&ti, &v)
		if err != nil {
			return DataFrame[uint64]{}, err
		}

		df.Timestamps = append(df.Timestamps, ti)
		df.Values = append(df.Values, v)
	}

	return df, nil
}
