package clickhouse

import (
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
)

type Ch struct {
	driver.Conn
}

// ProvideCh define a wire provider for Ch.
func ProvideCh(logger log.Logger, cfg config.Clickhouse) Ch {
	// Execute migrations.
	db := connectSql(logger, cfg, 5)
	migrate(logger, db)

	// Connect using native interface.
	conn := connect(logger, cfg, 5)

	return Ch{conn}
}
