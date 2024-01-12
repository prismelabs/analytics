package postgres

import (
	"database/sql"

	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
)

type Pg struct {
	*sql.DB
}

// ProvidePg define a wire provider for Pg.
func ProvidePg(logger log.Logger, cfg config.Postgres) Pg {
	db := connect(logger, cfg, 5)
	migrate(logger, db)

	return Pg{db}
}
