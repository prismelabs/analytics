package postgres

import (
	"io"
	"testing"

	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/prismelabs/prismeanalytics/internal/log"
	"github.com/prismelabs/prismeanalytics/internal/secret"
	"github.com/stretchr/testify/require"
)

func TestPostgresConnect(t *testing.T) {
	logger := log.NewLogger("test_postgres_connect", io.Discard, false)

	t.Run("NonExistentPostgresInstance", func(t *testing.T) {
		cfg := config.Postgres{
			Url: secret.New("postgres://foo:bar@down.localhost:5432/public?sslmode=disable"),
		}

		require.Panics(t, func() {
			_ = connect(logger, cfg, 1)
		})
	})

	// We're not testing a real connection to postgres in unit tests.
}
