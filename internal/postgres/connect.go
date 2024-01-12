package postgres

import (
	"database/sql"
	"time"

	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
)

// connect tries to connect to postgres instance using the given config.
// This function panics if `maxRetry` connect attempt fails.
func connect(logger log.Logger, cfg config.Postgres, maxRetry int) *sql.DB {
	var db *sql.DB
	var err error

	for retry := 0; retry < maxRetry; retry++ {
		logger.Info().
			Int("retry", retry).
			Int("max_retry", maxRetry).
			Msg("trying to connect to postgres")

		time.Sleep(time.Duration(retry) * time.Second)

		db, err = sql.Open("postgres", cfg.Url.ExposeSecret())
		if err != nil {
			continue
		}

		err = db.Ping()
		if err != nil {
			continue
		}

		break
	}

	if err != nil {
		logger.Panic().Msgf("failed to connect to postgres: %v", err.Error())
	}

	return db
}
