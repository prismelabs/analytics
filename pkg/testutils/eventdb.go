package testutils

import (
	"context"
	"testing"

	"github.com/prismelabs/analytics/pkg/sql"
	"github.com/stretchr/testify/require"
)

// DropTables drops all tables in `prisme` database.
func DropTables(t *testing.T, db sql.DB) {
	rows, err := db.Query(context.Background(), "SELECT name FROM system.tables WHERE database = 'prisme'")
	require.NoError(t, err, "failed to list prisme tables")

	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		require.NoError(t, err, "failed to scan table name")

		err = db.Exec(context.Background(), "DROP TABLE prisme."+table)
		require.NoErrorf(t, err, "failed to drop table prisme.%v", table)
	}
}
