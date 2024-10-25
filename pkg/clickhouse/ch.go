// Package clickhouse contains wire provider for clickhouse connection and
// migration.
package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/rs/zerolog"
)

// Ch define a ClickHouse OLAP engine interface.
type Ch interface {
	Select(ctx context.Context, dest any, query string, args ...any) error
	Query(ctx context.Context, query string, args ...any) (driver.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) driver.Row
	PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error)
	Exec(ctx context.Context, query string, args ...any) error
}

// ProvideClickhouse define a wire provider for clickhouse based Ch.
func ProvideClickhouse(logger zerolog.Logger, cfg config.Clickhouse, source source.Driver) Ch {
	// Execute migrations.
	sqlLogger := logger.With().
		Str("service", "clickhouse_provider").
		Str("engine", "clickhouse").
		Str("protocol", "http").
		Logger()
	db := connectSql(sqlLogger, cfg, 5)
	migrate(sqlLogger, db, cfg.Database, source)

	logger = logger.With().
		Str("service", "clickhouse_provider").
		Str("protocol", "native").
		Logger()

	// Connect using native interface.
	conn := Connect(logger, cfg, 5)

	return conn
}

// ProvideChDb define a wire provider for chdb based Ch.
func ProvideChDb(logger zerolog.Logger, cfg config.ChDb, source source.Driver) Ch {
	logger = logger.With().
		Str("service", "clickhouse_provider").
		Str("engine", "chdb").
		Logger()

	chdb, err := newChDb(cfg)
	if err != nil {
		logger.Panic().Msgf("failed to create chdb sql session: %v", err.Error())
	}

	// Execute migrations.
	migrate(logger, chdb.DB, "default", source)

	// Create another chdb
	chdb, err = newChDb(cfg)
	if err != nil {
		logger.Panic().Msgf("failed to create chdb sql session: %v", err.Error())
	}

	return chdb
}
