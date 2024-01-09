package config

// Postgres connection options.
type Postgres struct {
	Url SecretString
}

// PostgresFromEnv loads postgres config from environment variables.
// This function panics if required environment variables are missing.
func PostgresFromEnv() Postgres {
	return Postgres{
		Url: NewSecretString(mustGetEnv("PRISME_POSTGRES_URL")),
	}
}
