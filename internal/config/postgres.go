package config

import "github.com/prismelabs/prismeanalytics/internal/secret"

// Postgres related options.
type Postgres struct {
	Url secret.Secret[string]
}

// PostgresFromEnv loads postgres related options from environment variables.
// This function panics if required environment variables are missing.
func PostgresFromEnv() Postgres {
	return Postgres{
		Url: secret.New(mustGetEnv("PRISME_POSTGRES_URL")),
	}
}
