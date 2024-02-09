package clickhouse

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/prismelabs/prismeanalytics/pkg/config"
	"github.com/prismelabs/prismeanalytics/pkg/log"
)

// connect connects to clickhouse and returns a driver.Conn.
// This function panics if `maxRetry` connect attempt fails.
func connect(logger log.Logger, cfg config.Clickhouse, maxRetry int) (conn driver.Conn) {
	// Build clickhouse options.
	var clickHouseTls *tls.Config = nil
	if cfg.TlsEnabled {
		clickHouseTls = &tls.Config{}
	}
	options := clickhouse.Options{
		Addr: []string{cfg.HostPort},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User.ExposeSecret(),
			Password: cfg.Password.ExposeSecret(),
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "prismeanalytics.com", Version: "0.1"},
			},
		},
		Debugf: func(format string, v ...interface{}) {
			logger.Debug().Msgf(format, v...)
		},
		TLS: clickHouseTls,
	}

	// Try to connect.
	var err error
	for retry := 0; retry < maxRetry; retry++ {
		logger.Info().
			Int("retry", retry).
			Int("max_retry", maxRetry).
			Strs("clickhouse_addr", options.Addr).
			Msg("trying to establish native connection to clickhouse")

		time.Sleep(time.Duration(retry) * time.Second)

		conn, err = clickhouse.Open(&options)
		if err != nil {
			continue
		}

		err = conn.Ping(context.Background())
		if err != nil {
			continue
		}

		logger.Info().Msg("clickhouse native connection established")
		break
	}

	if err != nil {
		logger.Panic().Msgf("failed to connect to clickhouse: %v", err.Error())
	}

	return conn
}

// Connect connects to clickhouse and returns a sql.DB.
// This function panics if `maxRetry` connect attempt fails.
func connectSql(logger log.Logger, cfg config.Clickhouse, maxRetry int) *sql.DB {
	var db *sql.DB
	var err error

	for retry := 0; retry < maxRetry; retry++ {
		logger.Info().
			Int("retry", retry).
			Int("max_retry", maxRetry).
			Msg("trying to establish SQL connection to clickhouse")

		time.Sleep(time.Duration(retry) * time.Second)

		connectionString := fmt.Sprintf(
			"clickhouse://%v/%v?username=%v&password=%v",
			cfg.HostPort,
			cfg.Database,
			cfg.User.ExposeSecret(),
			cfg.Password.ExposeSecret(),
		)
		if cfg.TlsEnabled {
			connectionString += "&secure=true"
		}

		db, err = sql.Open("clickhouse", connectionString)
		if err != nil {
			println(err.Error())
			continue
		}

		err = db.Ping()
		if err != nil {
			println(err.Error())
			continue
		}

		logger.Info().Msg("clickhouse SQL connection established")
		break
	}

	if err != nil {
		logger.Panic().Msgf("failed to connect to clickhouse: %v", err.Error())
	}

	return db
}
