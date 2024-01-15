package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

func getEnvOrDefault(name string, defaultValue string) string {
	if env, envDefined := os.LookupEnv(name); envDefined {
		return env
	}

	return defaultValue
}

func mustGetEnv(name string) string {
	if env, envDefined := os.LookupEnv(name); envDefined {
		if env == "" {
			panic(fmt.Errorf("%v environment variable is an empty string", name))
		}

		return env
	}

	panic(fmt.Errorf("%v environment variable not set", name))
}

func parseUintEnvOrDefault(name string, defaultValue uint64, bitSize int) uint64 {
	if env, envDefined := os.LookupEnv(name); envDefined {
		value, err := strconv.ParseUint(env, 10, bitSize)
		if err != nil {
			panic(fmt.Errorf("%v environment is not a valid uint%d", name, bitSize))
		}

		return value
	}

	return defaultValue
}

func mustParseUrlEnv(name string) *url.URL {
	rawUrl := mustGetEnv(name)
	u, err := url.Parse(rawUrl)
	if err != nil {
		panic(fmt.Errorf("%v environment variable is not a valid URL", name))
	}

	return u
}
