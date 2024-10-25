package clickhouse

import (
	"context"
	"database/sql"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	_ "github.com/chdb-io/chdb-go/chdb/driver"
	"github.com/prismelabs/analytics/pkg/config"
)

var _ Ch = chDb{}

// chDb is a wrapper around *sql.Db that implements Ch.
// This enables Prisme to works using an embedded version of clickhouse.
type chDb struct {
	*sql.DB
}

func newChDb(cfg config.ChDb) (chDb, error) {
	db, err := sql.Open("chdb", "session="+cfg.Path)
	return chDb{db}, err
}

// Select implements driver.Conn.
func (cs chDb) Select(ctx context.Context, dest any, query string, args ...any) error {
	panic("not implemented")
}

// Query implements driver.Conn.
func (cs chDb) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	panic("not implemented")
}

// QueryRow implements driver.Conn.
func (cs chDb) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	panic("not implemented")
}

// PrepareBatch implements driver.Conn.
func (cs chDb) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	panic("not implemented")
}

// Exec implements driver.Conn.
func (cs chDb) Exec(ctx context.Context, query string, args ...any) error {
	_, err := cs.DB.ExecContext(ctx, query, args...)
	return err
}

// Close implements driver.Conn.
func (cs chDb) Close() error {
	return cs.DB.Close()
}
