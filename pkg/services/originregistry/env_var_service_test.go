package originregistry

import (
	"context"
	"io"
	"testing"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestEnvVarService(t *testing.T) {
	t.Run("IsOriginRegistered", func(t *testing.T) {
		ctx := context.Background()
		logger := log.New("env_var_service_test", io.Discard, false)

		t.Run("NonRegistered", func(t *testing.T) {
			service := NewService(Config{Origins: "notexample.com"}, logger)

			isRegistered, err := service.IsOriginRegistered(ctx, "example.com")
			require.NoError(t, err)
			require.False(t, isRegistered)
		})

		t.Run("Registered", func(t *testing.T) {
			service := NewService(Config{Origins: "example.org,example.com"}, logger)

			isRegistered, err := service.IsOriginRegistered(ctx, "example.com")
			require.NoError(t, err)
			require.True(t, isRegistered)
		})
	})
}
