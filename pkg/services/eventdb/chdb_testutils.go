//go:build test && chdb

package eventdb

import (
	"io"

	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/testutils"

	"github.com/stretchr/testify/require"
)

// NewChDb return a new chdb based Service instance for testing only.
func NewChDb(t require.TestingT) (Service, teardown.Service) {
	logger := log.New("chdb-test", io.Discard, false)
	teardown := teardown.NewService()
	var chdbCfg chdb.Config
	testutils.ConfigueLoad(t, &chdbCfg)
	db, err := NewService(
		Config{Driver: "chdb"},
		chdbCfg,
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
