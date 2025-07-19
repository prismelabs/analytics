package chdb

import (
	"database/sql"

	_ "github.com/chdb-io/chdb-go/chdb/driver"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/rs/zerolog"
)

type ChDb struct {
	*sql.DB
}

// NewChDb returns a new chdb session.
func NewChDb(
	logger zerolog.Logger,
	cfg Config,
	source source.Driver,
	teardown teardown.Service,
) ChDb {
	sqlLogger := logger.With().
		Str("service", "chdb_provider").
		Logger()
	db, err := sql.Open("chdb", "session="+cfg.Path)
	if err != nil {
		logger.Panic().Err(err).Msg("failed to open chdb based *sql.DB")
	}
	migrate(sqlLogger, db, source)

	teardown.RegisterProcedure(func() error {
		return db.Close()
	})

	return ChDb{db}
}
