//go:build test && chdb

package eventdb

import (
	"io"
	"testing"

	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/stretchr/testify/require"
)

// NewChDb return a new chdb based Service instance for testing only.
func NewChDb(tb testing.TB) (Service, teardown.Service) {
	logger := log.New("eventdb-chdb-testutils", io.Discard, false)
	teardown := teardown.NewService()
	var chdbCfg chdb.Config
	testutils.ConfigueLoad(tb, &chdbCfg)
	driver := clickhouse.EmbeddedSourceDriver(logger)
	db, err := NewService(
		Config{Driver: "chdb"},
		chdbCfg,
		logger,
		driver,
		teardown,
	)
	require.NoError(tb, err)
	teardown.RegisterProcedure(func() error {
		return testutils.DropTables(db)
	})

	return db, teardown
}

// NewClickHouse return a new ClickHouse based Service instance for testing
// only.
func NewClickHouse(tb testing.TB) (Service, teardown.Service) {
	logger := log.New("eventdb-clickhouse-testutils", io.Discard, false)
	teardown := teardown.NewService()
	var clickHouseCfg clickhouse.Config
	testutils.ConfigueLoad(tb, &clickHouseCfg)
	driver := clickhouse.EmbeddedSourceDriver(logger)
	db, err := NewService(
		Config{Driver: "clickhouse"},
		clickHouseCfg,
		logger,
		driver,
		teardown,
	)
	require.NoError(tb, err)
	teardown.RegisterProcedure(func() error {
		return testutils.DropTables(db)
	})

	return db, teardown
}

// ForEachDriver calls `cb` with all available Service implementations.
func ForEachDriver(tb testing.TB, cb func(Service)) {
	for driver := range Drivers() {
		switch driver {
		case "clickhouse":
			db, teardown := NewClickHouse(tb)
			cb(db)
			require.NoError(tb, teardown.Teardown())
		case "chdb":
			db, teardown := NewChDb(tb)
			cb(db)
			require.NoError(tb, teardown.Teardown())
		default:
			panic("unreachable")
		}
	}
}
