package grafana

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegClientHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cfg := configFromEnv()
	t.Run("Healthy", func(t *testing.T) {
		cli := ProvideClient(cfg)

		err := cli.HealthCheck(context.Background())
		require.NoError(t, err)
	})

	t.Run("NonExistentInstance", func(t *testing.T) {
		cfg := cfg
		cfg.Url = "http://down.localhost"
		cli := ProvideClient(cfg)

		err := cli.HealthCheck(context.Background())
		require.Error(t, err)
		require.Equal(t, "failed to query grafana for health check: error when dialing [::1]:80: dial tcp [::1]:80: connect: connection refused", err.Error())
	})
}
