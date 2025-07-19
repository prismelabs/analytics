package grafana

import (
	"context"
	"testing"

	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestIntegClientHealthCheck(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	t.Run("Healthy", func(t *testing.T) {
		cli := NewClient(cfg)

		err := cli.HealthCheck(context.Background())
		require.NoError(t, err)
	})

	t.Run("NonExistentInstance", func(t *testing.T) {
		cfg := cfg
		cfg.Url = "http://down.localhost"
		cli := NewClient(cfg)

		err := cli.HealthCheck(context.Background())
		require.Error(t, err)
		require.Equal(t, "failed to query grafana for health check: error when dialing [::1]:80: dial tcp [::1]:80: connect: connection refused", err.Error())
	})
}
