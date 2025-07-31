//go:build !race

package eventdb

import (
	"context"
	"io"
	"testing"

	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestIntegNoRaceDetectorService(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	setup := func(t *testing.T, driver string, driverCfg any) (Service, teardown.Service) {
		logger := log.New("eventdb_service_test", io.Discard, false)
		cfg := Config{
			Driver: driver,
		}
		source := clickhouse.EmbeddedSourceDriver(logger)
		teardown := teardown.NewService()

		db, err := NewService(cfg, driverCfg, logger, source, teardown)
		require.NoError(t, err)

		return db, teardown
	}

	forEachDriver := func(t *testing.T, fn func(t *testing.T, db Service)) {
		for driver := range Drivers() {
			var driverCfg any
			switch driver {
			case "chdb":
				var cfg chdb.Config
				testutils.ConfigueLoad(t, &cfg)
				driverCfg = cfg
			case "clickhouse":
				var cfg clickhouse.Config
				testutils.ConfigueLoad(t, &cfg)
				driverCfg = cfg
			}
			t.Run(driver, func(t *testing.T) {
				db, teardown := setup(t, driver, driverCfg)
				fn(t, db)
				require.NoError(t, teardown.Teardown())
			})
		}
	}

	t.Run("DriverName", func(t *testing.T) {
		drivers := 0
		forEachDriver(t, func(t *testing.T, db Service) {
			switch db.DriverName() {
			case "chdb":
				drivers |= 1
			case "clickhouse":
				drivers |= 2
			default:
				t.FailNow()
			}
		})

		require.Equal(t, 3, drivers)
	})

	t.Run("Driver", func(t *testing.T) {
		forEachDriver(t, func(t *testing.T, db Service) {
			switch db.DriverName() {
			case "chdb":
				require.IsType(t, chdb.ChDb{}, db.Driver())
			case "clickhouse":
				require.IsType(t, clickhouse.Ch{}, db.Driver())
			default:
				t.FailNow()
			}
		})
	})

	t.Run("QueryRow", func(t *testing.T) {
		t.Run("SimpleSelect", func(t *testing.T) {
			forEachDriver(t, func(t *testing.T, db Service) {
				row := db.QueryRow(context.Background(), "SELECT 1")
				var n uint8
				err := row.Scan(&n)
				require.NoError(t, err)
				require.Equal(t, uint8(1), n)
			})
		})
	})

	t.Run("Query", func(t *testing.T) {
		t.Run("SimpleSelect", func(t *testing.T) {
			forEachDriver(t, func(t *testing.T, db Service) {
				result, err := db.Query(context.Background(), "SELECT 1")
				require.NoError(t, err)

				require.True(t, result.Next())

				var n uint8
				err = result.Scan(&n)
				require.NoError(t, err)
				require.Equal(t, uint8(1), n)
			})
		})
	})

	t.Run("Exec", func(t *testing.T) {
		t.Run("CreateTable", func(t *testing.T) {
			forEachDriver(t, func(t *testing.T, db Service) {
				query := "CREATE TABLE IF NOT EXISTS foo(bar String) ENGINE = Memory"
				err := db.Exec(context.Background(), query)
				require.NoError(t, err)
			})
		})
		t.Run("DropTable", func(t *testing.T) {
			forEachDriver(t, func(t *testing.T, db Service) {
				query := "DROP TABLE IF EXISTS foo"
				err := db.Exec(context.Background(), query)
				require.NoError(t, err)
			})
		})
	})

}
