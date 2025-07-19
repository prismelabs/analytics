package clickhouse

import (
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/prismelabs/analytics/pkg/embedded"
	"github.com/prismelabs/analytics/pkg/log"
)

// EmbeddedSourceDriver returns golang-migrate source driver from embedded file
// system.
func EmbeddedSourceDriver(logger log.Logger) source.Driver {
	source, err := iofs.New(embedded.ChMigrations, "ch_migrations")
	logger.Fatal("failed to retrieve clickhouse migration source", err)

	return source
}
