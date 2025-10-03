//go:build test

package eventdb

import (
	"io"

	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/stretchr/testify/require"
)

// NewClickHouse return a new ClickHouse based Service instance for testing
// only.
func NewClickHouse(t require.TestingT) (Service, teardown.Service) {
	logger := log.New("clickhouse-test", io.Discard, false)
	teardown := teardown.NewService()
	var clickHouseCfg clickhouse.Config
	testutils.ConfigueLoad(t, &clickHouseCfg)
	db, err := NewService(
		Config{Driver: "clickhouse"},
		clickHouseCfg,
		logger,
		clickhouse.EmbeddedSourceDriver(logger),
		teardown,
	)
	require.NoError(t, err)
	teardown.RegisterProcedure(func() error {
		return testutils.DropTables(db)
	})

	return db, teardown
}
