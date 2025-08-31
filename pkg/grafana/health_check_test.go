package grafana

import (
	"context"
	"regexp"
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
		require.Regexp(t,
			regexp.MustCompile("failed to query grafana for health check: error when dialing [^ ]+ dial tcp [^ ]+ connect: connection refused"),
			err.Error(),
		)
	})
}
