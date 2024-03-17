package originregistry

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/prismelabs/analytics/pkg/log"
	"github.com/stretchr/testify/require"
)

func TestEnvVarService(t *testing.T) {
	t.Run("ProvideEnvVarService", func(t *testing.T) {
		os.Clearenv()
		t.Run("PRISME_ORIGIN_REGISTRY_ORIGINS/NotSet", func(t *testing.T) {
			require.Panics(t, func() {
				logger := log.NewLogger("env_var_service_test", io.Discard, false)
				ProvideEnvVarService(logger)
			})
		})

		os.Clearenv()
		t.Run("PRISME_ORIGIN_REGISTRY_ORIGINS/Set/EmptyString", func(t *testing.T) {
			require.Panics(t, func() {
				os.Setenv("PRISME_ORIGIN_REGISTRY_ORIGINS", "")
				logger := log.NewLogger("env_var_service_test", io.Discard, false)
				ProvideEnvVarService(logger)
			})
		})

		os.Clearenv()
		t.Run("PRISME_ORIGIN_REGISTRY_ORIGINS/Set/NonEmptyString", func(t *testing.T) {
			os.Setenv("PRISME_ORIGIN_REGISTRY_ORIGINS", "example.com")
			logger := log.NewLogger("env_var_service_test", io.Discard, false)
			ProvideEnvVarService(logger)
		})

		t.Run("Depreacted", func(t *testing.T) {
			os.Clearenv()
			t.Run("PRISME_SOURCE_REGISTRY_SOURCES/NotSet", func(t *testing.T) {
				require.Panics(t, func() {
					logger := log.NewLogger("env_var_service_test", io.Discard, false)
					ProvideEnvVarService(logger)
				})
			})

			os.Clearenv()
			t.Run("PRISME_SOURCE_REGISTRY_SOURCES/Set/EmptyString", func(t *testing.T) {
				require.Panics(t, func() {
					os.Setenv("PRISME_SOURCE_REGISTRY_SOURCES", "")
					logger := log.NewLogger("env_var_service_test", io.Discard, false)
					ProvideEnvVarService(logger)
				})
			})

			os.Clearenv()
			t.Run("PRISME_SOURCE_REGISTRY_SOURCES/Set/NonEmptyString", func(t *testing.T) {
				os.Setenv("PRISME_SOURCE_REGISTRY_SOURCES", "example.com")
				logger := log.NewLogger("env_var_service_test", io.Discard, false)
				ProvideEnvVarService(logger)
			})

			t.Run("PRISME_SOURCE_REGISTRY_SOURCES/PRISME_ORIGIN_REGISTRY_ORIGINS/Set/NonEmptyString", func(t *testing.T) {
				require.Panics(t, func() {
					os.Setenv("PRISME_SOURCE_REGISTRY_SOURCES", "example.com")
					os.Setenv("PRISME_ORIGIN_REGISTRY_ORIGINS", "example.com")
					logger := log.NewLogger("env_var_service_test", io.Discard, false)
					ProvideEnvVarService(logger)
				})
			})
		})
	})

	t.Run("IsOriginRegistered", func(t *testing.T) {
		ctx := context.Background()
		logger := log.NewLogger("env_var_service_test", io.Discard, false)

		os.Clearenv()
		t.Run("NonRegistered", func(t *testing.T) {
			os.Setenv("PRISME_SOURCE_REGISTRY_SOURCES", "notexample.com")
			service := ProvideEnvVarService(logger)

			isRegistered, err := service.IsOriginRegistered(ctx, "example.com")
			require.NoError(t, err)
			require.False(t, isRegistered)
		})

		os.Clearenv()
		t.Run("Registered", func(t *testing.T) {
			os.Setenv("PRISME_SOURCE_REGISTRY_SOURCES", "example.org,example.com")

			service := ProvideEnvVarService(logger)

			isRegistered, err := service.IsOriginRegistered(ctx, "example.com")
			require.NoError(t, err)
			require.True(t, isRegistered)
		})
	})
}
