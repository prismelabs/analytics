package clickhouse

import (
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/prismelabs/analytics/pkg/embedded"
	"github.com/rs/zerolog"
)

// ProvideEmbeddedSourceDriver is a wire provider for golang-migrate source driver.
func ProvideEmbeddedSourceDriver(logger zerolog.Logger) source.Driver {
	source, err := iofs.New(embedded.ChMigrations, "ch_migrations")
	if err != nil {
		logger.Panic().Msgf("failed to retrieve clickhouse migration source: %v", err.Error())
	}

	return source
}
