// Package clickhouse contains wire provider for clickhouse connection and
// migration.
package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/rs/zerolog"
)

// Ch define a connection to a ClickHouse instance.
type Ch struct {
	driver.Conn
}

// ProvideCh define a wire provider for Ch.
func ProvideCh(
	logger zerolog.Logger,
	cfg Config,
	source source.Driver,
	teardown teardown.Service,
) Ch {

	// Execute migrations.
	sqlLogger := logger.With().
		Str("service", "clickhouse_provider").
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

	// Close connection on teardown.
	teardown.RegisterProcedure(func() error {
		return conn.Close()
	})

	return Ch{conn}
}
