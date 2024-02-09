package clickhouse

import (
	"io"
	"testing"

	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/secret"
	"github.com/stretchr/testify/require"
)

func TestClickhouseConnect(t *testing.T) {
	logger := log.NewLogger("test_clickhouse_connect", io.Discard, false)

	t.Run("NonExistentInstance", func(t *testing.T) {
		cfg := config.Clickhouse{
			TlsEnabled: false,
			HostPort:   "down.localhost",
			Database:   "analytics",
			User:       secret.New("foo"),
			Password:   secret.New("bar"),
		}

		require.Panics(t, func() {
			_ = connect(logger, cfg, 1)
		})
	})

	// We're not testing a real connection to postgres in unit tests.
}

func TestClickhouseConnectSql(t *testing.T) {
	logger := log.NewLogger("test_clickhouse_connect_sql", io.Discard, false)

	t.Run("NonExistentInstance", func(t *testing.T) {
		cfg := config.Clickhouse{
			TlsEnabled: false,
			HostPort:   "down.localhost",
			Database:   "analytics",
			User:       secret.New("foo"),
			Password:   secret.New("bar"),
		}

		require.Panics(t, func() {
			_ = connectSql(logger, cfg, 1)
		})
	})

	// We're not testing a real connection to postgres in unit tests.
}
