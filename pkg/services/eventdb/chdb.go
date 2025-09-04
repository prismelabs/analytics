//go:build chdb

package eventdb

import (
	"context"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/sql"
)

func init() {
	dbFactory["chdb"] = newChDb
}

type chDb struct {
	chdb chdb.ChDb
}

func newChDb(logger log.Logger, cfg any, source source.Driver, teardown teardown.Service) (Service, error) {
	chdb := chdb.NewChDb(logger, cfg.(chdb.Config), source, teardown)
	return &chDb{chdb: chdb}, nil
}

// Exec implements Service.
func (c *chDb) Exec(ctx context.Context, query string, args ...any) error {
	_, err := c.chdb.ExecContext(ctx, query, args...)
	return err
}

// Query implements Service.
func (c *chDb) Query(ctx context.Context, query string, args ...any) (sql.QueryResult, error) {
	return c.chdb.QueryContext(ctx, query, args...)
}

// QueryRow implements Service.
func (c *chDb) QueryRow(ctx context.Context, query string, args ...any) sql.Row {
	return c.chdb.QueryRowContext(ctx, query, args...)
}

// DriverName implements Service.
func (c *chDb) DriverName() string {
	return "chdb"
}

// Driver implements Service.
func (c *chDb) Driver() any {
	return c.chdb
}
