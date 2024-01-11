package config

import "github.com/prismelabs/prismeanalytics/internal/secretstring"

// Postgres related options.
type Postgres struct {
	Url secretstring.SecretString
}

// PostgresFromEnv loads postgres related options from environment variables.
// This function panics if required environment variables are missing.
func PostgresFromEnv() Postgres {
	return Postgres{
		Url: secretstring.NewSecretString(mustGetEnv("PRISME_POSTGRES_URL")),
	}
}
