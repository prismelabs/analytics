package clickhouse

import (
	"database/sql"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/rs/zerolog"
)

// migrate starts migrating a clickhouse instance to the latest version.
func migrate(logger zerolog.Logger, db *sql.DB, dbName string, source source.Driver) {
	driverCfg := clickhouse.Config{
		DatabaseName:          dbName,
		MigrationsTable:       "migrations",
		MigrationsTableEngine: "MergeTree",
		MultiStatementEnabled: true,
	}

	driver, err := clickhouse.WithInstance(db, &driverCfg)
	if err != nil {
		logger.Panic().Msgf("failed to create golang-migrate driver for clickhouse migration: %v", err.Error())
	}

	m, err := gomigrate.NewWithInstance("migrations", source, dbName, driver)
	if err != nil {
		logger.Panic().Msgf("failed to create go-migrate.Migrate instance: %v", err.Error())
	}
	m.Log = log.GoMigrateLogger(logger)

	err = m.Up()
	if err != nil && err != gomigrate.ErrNoChange {
		logger.Panic().Msgf("failed to execute clickhouse migrations: %v", err.Error())
	}

	srcErr, driErr := m.Close()
	if srcErr != nil || driErr != nil {
		logger.Panic().
			AnErr("source_error", srcErr).
			AnErr("driver_error", driErr).
			Msg("failed to close migration source and/or migration sql driver")
	}

	logger.Info().Msg("clickhouse migration successfully done")
}
