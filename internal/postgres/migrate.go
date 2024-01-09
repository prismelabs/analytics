package postgres

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/prismelabs/prismeanalytics/internal/embedded"
	"github.com/prismelabs/prismeanalytics/internal/log"
)

// Migrate starts migrating a postgres instance to the latest version.
func Migrate(logger log.Logger, db *sql.DB) {
	source, err := iofs.New(embedded.PgMigrations, "pg_migrations")
	if err != nil {
		logger.Panic().Msgf("failed to retrieve postgres migration source: %v", err.Error())
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable:       "migrations",
		MigrationsTableQuoted: false,
		MultiStatementEnabled: false,
		DatabaseName:          "analytics",
		SchemaName:            "public",
		StatementTimeout:      0,
		MultiStatementMaxSize: 0,
	})
	if err != nil {
		logger.Panic().Msgf("failed to create golang-migrate driver for postgres migration: %v", err.Error())
	}

	m, err := migrate.NewWithInstance("migrations", source, "analytics", driver)
	if err != nil {
		logger.Panic().Msgf("failed to create go-migrate.Migrate instance: %v", err.Error())
	}
	m.Log = &logger

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Panic().Msgf("failed to execute postgres migrations: %v", err.Error())
	}

	logger.Info().Msg("postgres migration successfully done")
}
