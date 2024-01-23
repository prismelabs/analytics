package sourceregistry

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/prismelabs/prismeanalytics/internal/log"
	"github.com/stretchr/testify/require"
)

type staticSource struct {
	value string
}

// SourceName implements Source.
func (ss staticSource) SourceString() string {
	return ss.value
}

func TestEnvVarService(t *testing.T) {
	t.Run("ProvideEnvVarService", func(t *testing.T) {
		t.Run("PRISME_SOURCE_REGISTRY_SOURCES/NotSet", func(t *testing.T) {
			require.Panics(t, func() {
				logger := log.NewLogger("env_var_service_test", io.Discard, false)
				ProvideEnvVarService(logger)
			})
		})

		t.Run("PRISME_SOURCE_REGISTRY_SOURCES/Set/EmptyString", func(t *testing.T) {
			require.Panics(t, func() {
				os.Setenv("PRISME_SOURCE_REGISTRY_SOURCES", "")
				logger := log.NewLogger("env_var_service_test", io.Discard, false)
				ProvideEnvVarService(logger)
			})
		})

		t.Run("PRISME_SOURCE_REGISTRY_SOURCES/Set/NonEmptyString", func(t *testing.T) {
			os.Setenv("PRISME_SOURCE_REGISTRY_SOURCES", "example.com")
			logger := log.NewLogger("env_var_service_test", io.Discard, false)
			ProvideEnvVarService(logger)
		})
	})

	t.Run("IsSourceRegistered", func(t *testing.T) {
		ctx := context.Background()
		logger := log.NewLogger("env_var_service_test", io.Discard, false)

		t.Run("NonRegistered", func(t *testing.T) {
			os.Setenv("PRISME_SOURCE_REGISTRY_SOURCES", "notexample.com")
			service := ProvideEnvVarService(logger)

			isRegistered, err := service.IsSourceRegistered(ctx, staticSource{"example.com"})
			require.NoError(t, err)
			require.False(t, isRegistered)
		})

		t.Run("Registered", func(t *testing.T) {
			os.Setenv("PRISME_SOURCE_REGISTRY_SOURCES", "example.org,example.com")

			service := ProvideEnvVarService(logger)

			isRegistered, err := service.IsSourceRegistered(ctx, staticSource{"example.com"})
			require.NoError(t, err)
			require.True(t, isRegistered)
		})
	})
}
