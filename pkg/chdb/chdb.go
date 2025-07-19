package chdb

import (
	"database/sql"

	_ "github.com/chdb-io/chdb-go/chdb/driver"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/teardown"
)

type ChDb struct {
	*sql.DB
}

// NewChDb returns a new chdb session.
func NewChDb(
	logger log.Logger,
	cfg Config,
	source source.Driver,
	teardown teardown.Service,
) ChDb {
	sqlLogger := logger.With(
		"service", "chdb_provider",
	)
	db, err := sql.Open("chdb", "session="+cfg.Path)
	if err != nil {
		logger.Fatal("failed to open chdb based *sql.DB", err)
	}
	migrate(sqlLogger, db, source)

	teardown.RegisterProcedure(func() error {
		return db.Close()
	})

	return ChDb{db}
}
