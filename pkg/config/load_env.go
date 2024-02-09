package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

// GetEnvOrDefault reads environment variable with the given name and return it.
// If variable is not set, defaultValue is returned.
func GetEnvOrDefault(name string, defaultValue string) string {
	if env, envDefined := os.LookupEnv(name); envDefined {
		return env
	}

	return defaultValue
}

// MustGetEnv reads environment variable with the given name and return it.
// If variable is not set, this function panics.
func MustGetEnv(name string) string {
	if env, envDefined := os.LookupEnv(name); envDefined {
		if env == "" {
			panic(fmt.Errorf("%v environment variable is an empty string", name))
		}

		return env
	}

	panic(fmt.Errorf("%v environment variable not set", name))
}

// ParseUintEnvOrDefault reads environment variable with the given name and parses it
// as an unsigned integer.
// If variable is not set, defaultValue is returned.
// If variable value is not a valid unsigned integer, this function panics.
func ParseUintEnvOrDefault(name string, defaultValue uint64, bitSize int) uint64 {
	if env, envDefined := os.LookupEnv(name); envDefined {
		value, err := strconv.ParseUint(env, 10, bitSize)
		if err != nil {
			panic(fmt.Errorf("%v environment is not a valid uint%d", name, bitSize))
		}

		return value
	}

	return defaultValue
}

// ParseIntEnvOrDefault reads environment variable with the given name and parses it
// as an signed integer.
// If variable is not set, defaultValue is returned.
// If variable value is not a valid signed integer, this function panics.
func ParseIntEnvOrDefault(name string, defaultValue int64, bitSize int) int64 {
	if env, envDefined := os.LookupEnv(name); envDefined {
		value, err := strconv.ParseInt(env, 10, bitSize)
		if err != nil {
			panic(fmt.Errorf("%v environment is not a valid int%d", name, bitSize))
		}

		return value
	}

	return defaultValue
}

// MustParseUrlEnv reads and environment variable with the given name and
// parses it as an URL.
// If variable is not set, this function panics.
// If variable value is not a valid URL, this function panics.
func MustParseUrlEnv(name string) *url.URL {
	rawUrl := MustGetEnv(name)
	u, err := url.Parse(rawUrl)
	if err != nil {
		panic(fmt.Errorf("%v environment variable is not a valid URL", name))
	}

	return u
}

// ParseUintEnvOrDefault reads environment variable with the given name and parses it
// as a duration.
// If variable is not set, defaultValue is returned.
// If variable value is not a valid duration, this function panics.
func ParseDurationEnvOrDefault(name string, defaultValue time.Duration) time.Duration {
	if env, envDefined := os.LookupEnv(name); envDefined {
		value, err := time.ParseDuration(env)
		if err != nil {
			panic(fmt.Errorf("%v environment is not a valid duration", name))
		}

		return value
	}

	return defaultValue
}
