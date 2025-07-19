// Package clickhouse contains wire provider for clickhouse connection and
// migration.
package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/teardown"
)

// Ch define a connection to a ClickHouse instance.
type Ch struct {
	driver.Conn
}

// NewCh returns a new Ch object.
func NewCh(
	logger log.Logger,
	cfg Config,
	source source.Driver,
	teardown teardown.Service,
) Ch {

	// Execute migrations.
	sqlLogger := logger.With(
		"service", "clickhouse_provider",
		"protocol", "http",
	)
	db := connectSql(sqlLogger, cfg, 5)
	migrate(sqlLogger, db, cfg.Database, source)

	logger = logger.With(
		"service", "clickhouse_provider",
		"protocol", "native",
	)

	// Connect using native interface.
	conn := Connect(logger, cfg, 5)

	// Close connection on teardown.
	teardown.RegisterProcedure(func() error {
		return conn.Close()
	})

	return Ch{conn}
}
