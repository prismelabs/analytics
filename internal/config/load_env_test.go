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

func TestParseUintEnvOrDefault(t *testing.T) {
	os.Clearenv()
	t.Run("UndefinedVar", func(t *testing.T) {
		expected := uint64(42)
		actual := parseUintEnvOrDefault("MY_ENV_VAR", expected, 64)

		require.Equal(t, expected, actual)
	})

	os.Clearenv()
	t.Run("DefinedVar/NaN", func(t *testing.T) {
		os.Setenv("MY_ENV_VAR", "NaN")
		require.Panics(t, func() {
			parseUintEnvOrDefault("MY_ENV_VAR", 42, 64)
		})
	})

	os.Clearenv()
	t.Run("DefinedVar/ValidUint", func(t *testing.T) {
		os.Setenv("MY_ENV_VAR", "16")
		actual := parseUintEnvOrDefault("MY_ENV_VAR", 42, 64)
		require.Equal(t, uint64(16), actual)
	})
}
