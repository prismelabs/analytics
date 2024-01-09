package config

type Config struct {
	Server   Server
	Postgres Postgres
}

func FromEnv() Config {
	return Config{
		Server:   ServerFromEnv(),
		Postgres: PostgresFromEnv(),
	}
}
