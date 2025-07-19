package clickhouse

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/retry"
)

// Connect connects to clickhouse and returns a driver.Conn.
// This function panics if `maxRetry` Connect attempt fails.
func Connect(logger log.Logger, cfg Config, maxRetry uint) (conn driver.Conn) {
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
				{Name: "prismeanalytics"},
			},
		},
		MaxIdleConns: 16,
		Debugf: func(format string, v ...any) {
			logger.Debug(fmt.Sprintf(format, v...))
		},
		TLS: clickHouseTls,
	}

	// Try to connect.
	var err error
	err = retry.LinearBackoff(uint(maxRetry), time.Second, func(retry uint) error {
		logger.Info(
			"trying to establish native connection to clickhouse",
			"retry", retry,
			"max_retry", maxRetry,
			"clickhouse_addr", options.Addr,
		)

		time.Sleep(time.Duration(retry) * time.Second)

		conn, err = clickhouse.Open(&options)
		if err != nil {
			logger.Err("connection failed", err)
			return err
		}

		err = conn.Ping(context.Background())
		if err != nil {
			logger.Err("ping failed", err)
			return err
		}

		logger.Info("clickhouse native connection established")
		return nil
	}, retry.NeverCancel)
	if err != nil {
		logger.Fatal("failed to connect to clickhouse", err)
	}

	return conn
}

// Connect connects to clickhouse and returns a sql.DB.
// This function panics if `maxRetry` connect attempt fails.
func connectSql(logger log.Logger, cfg Config, maxRetry int) *sql.DB {
	var db *sql.DB
	var err error

	err = retry.LinearBackoff(uint(maxRetry), time.Second, func(retry uint) error {
		logger.Info(
			"trying to establish SQL connection to clickhouse",
			"retry", retry,
			"max_retry", maxRetry,
		)

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
			logger.Err("connection failed", err)
			return err
		}

		err = db.Ping()
		if err != nil {
			logger.Err("ping failed", err)
			return err
		}

		logger.Info("clickhouse SQL connection established")
		return nil
	}, retry.NeverCancel)
	if err != nil {
		logger.Fatal("failed to connect to clickhouse", err)
	}

	return db
}
