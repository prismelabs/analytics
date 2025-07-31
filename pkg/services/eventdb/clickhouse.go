package eventdb

import (
	"context"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/teardown"
)

func init() {
	dbFactory["clickhouse"] = newClickhouseDb
}

type clickhouseDb struct {
	ch clickhouse.Ch
}

func newClickhouseDb(logger log.Logger, cfg any, source source.Driver, teardown teardown.Service) (Service, error) {
	ch := clickhouse.NewCh(logger, cfg.(clickhouse.Config), source, teardown)
	return &clickhouseDb{ch}, nil
}

// Exec implements Service.
func (c *clickhouseDb) Exec(ctx context.Context, query string, args ...any) error {
	return c.ch.Exec(ctx, query, args...)
}

// Query implements Service.
func (c *clickhouseDb) Query(ctx context.Context, query string, args ...any) (QueryResult, error) {
	return c.ch.Query(ctx, query, args...)
}

// QueryRow implements Service.
func (c *clickhouseDb) QueryRow(ctx context.Context, query string, args ...any) Row {
	return c.ch.QueryRow(ctx, query, args...)
}

// DriverName implements Service.
func (c *clickhouseDb) DriverName() string {
	return "clickhouse"
}

// Driver implements Service.
func (c *clickhouseDb) Driver() any {
	return c.ch
}
