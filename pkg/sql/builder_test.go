package sql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	t.Run("SimpleQuery", func(t *testing.T) {
		query, args := (&Builder{}).
			Str("SELECT * FROM foo WHERE abc = ?", "123", 456).
			Strs("AND 1 != 2", "AND 2 != 3").
			Finish()
		require.Equal(t, "SELECT * FROM foo WHERE abc = ? AND 1 != 2 AND 2 != 3", query)
		require.Equal(t, []any{"123", 456}, args)
	})
}
