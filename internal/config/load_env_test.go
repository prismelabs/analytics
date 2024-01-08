package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEnvOrDefault(t *testing.T) {
	os.Clearenv()

	t.Run("UndefinedVar", func(t *testing.T) {
		expected := "MY_ENV_VAR_VALUE"
		actual := getEnvOrDefault("MY_ENV_VAR", expected)

		require.Equal(t, expected, actual)
	})

	os.Clearenv()
	t.Run("DefinedVar", func(t *testing.T) {
		expected := "MY_ENV_VAR_VALUE"
		os.Setenv("MY_ENV_VAR", expected)

		actual := getEnvOrDefault("MY_ENV_VAR", "")

		require.Equal(t, expected, actual)
	})
}

func TestMustGetEnv(t *testing.T) {
	os.Clearenv()
	t.Run("UndefinedVar", func(t *testing.T) {
		require.Panics(t, func() {
			mustGetEnv("MY_ENV_VAR")
		})
	})

	os.Clearenv()
	t.Run("DefinedVar", func(t *testing.T) {
		expected := "MY_ENV_VAR_VALUE"
		os.Setenv("MY_ENV_VAR", expected)

		actual := mustGetEnv("MY_ENV_VAR")

		require.Equal(t, expected, actual)
	})
}
