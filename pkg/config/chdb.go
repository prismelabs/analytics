package config

// ChDb (embedded Clickhouse SQL OLAP engine) options.
type ChDb struct {
	Path string
}

// ChDbFromEnv loads Chdb config from environment variables.
func ChDbFromEnv() ChDb {
	return ChDb{
		Path: GetEnvOrDefault("PRISME_CHDB_PATH", "./prisme_chdb"),
	}
}
