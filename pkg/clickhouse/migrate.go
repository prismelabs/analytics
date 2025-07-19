package clickhouse

import (
	"database/sql"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/log"
)

// migrate starts migrating a clickhouse instance to the latest version.
func migrate(logger log.Logger, db *sql.DB, dbName string, source source.Driver) {
	driver, err := clickhouse.WithInstance(db, &clickhouse.Config{
		DatabaseName:          dbName,
		MigrationsTable:       "migrations",
		MigrationsTableEngine: "MergeTree",
		MultiStatementEnabled: true,
	})
	logger.Fatal("failed to create golang-migrate driver for clickhouse migration", err)

	m, err := gomigrate.NewWithInstance("migrations", source, dbName, driver)
	logger.Fatal("failed to create go-migrate.Migrate instance", err)
	m.Log = log.GoMigrateLogger(logger)

	err = m.Up()
	if err != nil && err != gomigrate.ErrNoChange {
		logger.Fatal("failed to execute clickhouse migrations", err)
	}

	logger.Info("clickhouse migration successfully done")
}
