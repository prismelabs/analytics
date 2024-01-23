package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetEnvOrDefault(t *testing.T) {
	os.Clearenv()

	t.Run("UndefinedVar", func(t *testing.T) {
		expected := "MY_ENV_VAR_VALUE"
		actual := GetEnvOrDefault("MY_ENV_VAR", expected)

		require.Equal(t, expected, actual)
	})

	os.Clearenv()
	t.Run("DefinedVar", func(t *testing.T) {
		expected := "MY_ENV_VAR_VALUE"
		os.Setenv("MY_ENV_VAR", expected)

		actual := GetEnvOrDefault("MY_ENV_VAR", "")

		require.Equal(t, expected, actual)
	})
}

func TestMustGetEnv(t *testing.T) {
	os.Clearenv()
	t.Run("UndefinedVar", func(t *testing.T) {
		require.Panics(t, func() {
			MustGetEnv("MY_ENV_VAR")
		})
	})

	os.Clearenv()
	t.Run("DefinedVar", func(t *testing.T) {
		expected := "MY_ENV_VAR_VALUE"
		os.Setenv("MY_ENV_VAR", expected)

		actual := MustGetEnv("MY_ENV_VAR")

		require.Equal(t, expected, actual)
	})
}

func TestParseUintEnvOrDefault(t *testing.T) {
	os.Clearenv()
	t.Run("UndefinedVar", func(t *testing.T) {
		expected := uint64(42)
		actual := ParseUintEnvOrDefault("MY_ENV_VAR", expected, 64)

		require.Equal(t, expected, actual)
	})

	os.Clearenv()
	t.Run("DefinedVar/NaN", func(t *testing.T) {
		os.Setenv("MY_ENV_VAR", "NaN")
		require.Panics(t, func() {
			ParseUintEnvOrDefault("MY_ENV_VAR", 42, 64)
		})
	})

	os.Clearenv()
	t.Run("DefinedVar/ValidUint", func(t *testing.T) {
		os.Setenv("MY_ENV_VAR", "16")
		actual := ParseUintEnvOrDefault("MY_ENV_VAR", 42, 64)
		require.Equal(t, uint64(16), actual)
	})
}

func TestMustParseUrlEnv(t *testing.T) {
	os.Clearenv()
	t.Run("UndefinedVar", func(t *testing.T) {
		require.Panics(t, func() {
			MustParseUrlEnv("MY_ENV_VAR")
		})
	})

	os.Clearenv()
	t.Run("DefinedVar/InvalidUrl", func(t *testing.T) {
		os.Setenv("MY_ENV_VAR", "")
		require.Panics(t, func() {
			MustParseUrlEnv("MY_ENV_VAR")
		})
	})

	os.Clearenv()
	t.Run("DefinedVar/ValidUrl", func(t *testing.T) {
		expected := "http://admin:password@example.com:443/path#fragment?query=q"
		os.Setenv("MY_ENV_VAR", expected)
		actual := MustParseUrlEnv("MY_ENV_VAR")
		require.NotNil(t, actual)
		require.Equal(t, expected, actual.String())
	})
}

func TestParseDurationEnvOrDefault(t *testing.T) {
	os.Clearenv()
	t.Run("UndefinedVar", func(t *testing.T) {
		expected := time.Second
		actual := ParseDurationEnvOrDefault("MY_ENV_VAR", expected)

		require.Equal(t, expected, actual)
	})

	t.Run("DefinedVar/InvalidDuration", func(t *testing.T) {
		os.Setenv("MY_ENV_VAR", "not a duration")
		require.Panics(t, func() {
			_ = ParseDurationEnvOrDefault("MY_ENV_VAR", time.Second)
		})
	})

	os.Clearenv()
	t.Run("DefinedVar/ValidDuration", func(t *testing.T) {
		expected := "16s"
		os.Setenv("MY_ENV_VAR", expected)
		actual := ParseDurationEnvOrDefault("MY_ENV_VAR", time.Second)
		require.Equal(t, expected, actual.String())
	})
}
