package clickhouse

import (
	"io"
	"runtime"
	"testing"

	"github.com/negrel/secrecy"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestClickhouseConnect(t *testing.T) {
	logger := log.New("test_clickhouse_connect", io.Discard, false)

	t.Run("NonExistentInstance", func(t *testing.T) {
		cfg := Config{
			TlsEnabled: false,
			HostPort:   "down.localhost",
			Database:   "analytics",
			User:       secrecy.NewSecretString(secrecy.UnsafeStringToBytes("foo")),
			Password:   secrecy.NewSecretString(secrecy.UnsafeStringToBytes("bar")),
		}

		require.Panics(t, func() {
			_ = Connect(logger, cfg, 1)
		})
	})

	runtime.GC()
	// We're not testing a real connection to clickhouse in unit tests.
}

func TestClickhouseConnectSql(t *testing.T) {
	logger := log.New("test_clickhouse_connect_sql", io.Discard, false)

	t.Run("NonExistentInstance", func(t *testing.T) {
		cfg := Config{
			TlsEnabled: false,
			HostPort:   "down.localhost",
			Database:   "analytics",
			User:       secrecy.NewSecretString(secrecy.UnsafeStringToBytes("foo")),
			Password:   secrecy.NewSecretString(secrecy.UnsafeStringToBytes("bar")),
		}

		require.Panics(t, func() {
			_ = connectSql(logger, cfg, 1)
		})
	})

	runtime.GC()

	// We're not testing a real connection to clickhouse in unit tests.
}
