package clickhouse

import (
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/prismelabs/analytics/pkg/embedded"
	"github.com/rs/zerolog"
)

// EmbeddedSourceDriver returns golang-migrate source driver from embedded file
// system.
func EmbeddedSourceDriver(logger zerolog.Logger) source.Driver {
	source, err := iofs.New(embedded.ChMigrations, "ch_migrations")
	if err != nil {
		logger.Panic().Msgf("failed to retrieve clickhouse migration source: %v", err.Error())
	}

	return source
}
