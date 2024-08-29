package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/config"
	"github.com/rs/zerolog"
)

type Ch struct {
	driver.Conn
}

// ProvideCh define a wire provider for Ch.
func ProvideCh(logger zerolog.Logger, cfg config.Clickhouse, source source.Driver) Ch {

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

	return Ch{conn}
}
