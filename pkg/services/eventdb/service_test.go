//go:build test && !race && chdb

package eventdb

import (
	"context"
	"testing"

	"github.com/prismelabs/analytics/pkg/chdb"
	"github.com/prismelabs/analytics/pkg/clickhouse"
	"github.com/stretchr/testify/require"
)

func TestIntegNoRaceDetectorService(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	forEachDriver := func(t *testing.T, fn func(t *testing.T, db Service)) {
		t.Run("ClickHouse", func(t *testing.T) {
			db, teardown := NewClickHouse(t)
			fn(t, db)
			require.NoError(t, teardown.Teardown())
		})
		t.Run("ChDb", func(t *testing.T) {
			db, teardown := NewChDb(t)
			fn(t, db)
			require.NoError(t, teardown.Teardown())
		})
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
