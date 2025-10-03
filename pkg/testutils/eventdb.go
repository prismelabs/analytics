package testutils

import (
	"context"

	"github.com/prismelabs/analytics/pkg/sql"
)

// DropTables drops all tables in `prisme` database.
func DropTables(db sql.DB) error {
	rows, err := db.Query(context.Background(), "SELECT name FROM system.tables WHERE database = 'prisme'")
	if err != nil {
		return err
	}

	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			return err
		}

		err = db.Exec(context.Background(), "DROP TABLE prisme."+table)
		if err != nil {
			return err
		}
	}

	return nil
}
