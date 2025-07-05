package chdb

import (
	"database/sql"
	"time"

	_ "github.com/chdb-io/chdb-go/chdb/driver"
	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/rs/zerolog"
)

func migrate(logger zerolog.Logger, db *sql.DB, source source.Driver) {
	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS prisme")
	if err != nil {
		logger.Panic().Msgf("failed to create prisme database: %v", err.Error())
	}
	_, err = db.Exec("USE prisme")
	if err != nil {
		logger.Panic().Msgf("failed to use prisme database: %v", err.Error())
	}

	driver, err := clickhouse.WithInstance(db, &clickhouse.Config{
		DatabaseName:          "prisme",
		MigrationsTable:       "migrations",
		MigrationsTableEngine: "MergeTree",
		MultiStatementEnabled: true,
	})
	if err != nil {
		logger.Panic().Msgf("failed to create golang-migrate driver for clickhouse migration: %v", err.Error())
	}

	m, err := gomigrate.NewWithInstance("migrations", source, "prisme", driverWrapper{driver, db})
	if err != nil {
		logger.Panic().Msgf("failed to create go-migrate.Migrate instance: %v", err.Error())
	}
	m.Log = log.GoMigrateLogger(logger)

	err = m.Up()
	if err != nil && err != gomigrate.ErrNoChange {
		logger.Panic().Msgf("failed to execute chdb migrations: %v", err.Error())
	}

	logger.Info().Msg("chdb migration successfully done")
}

// A wrapper around golang-migrate clickhouse driver that overwrites SetVersion()
// as it uses sql.DB.Begin() instead of sql.DB.Exec.
type driverWrapper struct {
	database.Driver
	db *sql.DB
}

func (dw driverWrapper) SetVersion(version int, dirty bool) error {
	var (
		bool = func(v bool) uint8 {
			if v {
				return 1
			}
			return 0
		}
	)

	query := "INSERT INTO migrations (version, dirty, sequence) VALUES (?, ?, ?)"
	if _, err := dw.db.Exec(query, version, bool(dirty), time.Now().UnixNano()); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}

	return nil
}
