package config

import (
	"fmt"
	"os"
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
