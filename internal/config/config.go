package config

// Config contains all configuration options.
type Config struct {
	Server     Server
	Postgres   Postgres
	Clickhouse Clickhouse
}

// FromEnv builds a Config struct from environment variables.
func FromEnv() Config {
	return Config{
		Server:     ServerFromEnv(),
		Postgres:   PostgresFromEnv(),
		Clickhouse: ClickhouseFromEnv(),
	}
}
