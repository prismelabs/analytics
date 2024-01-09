package config

// Server specific options.
type Server struct {
	// Sets log level to debug.
	Debug bool
	// Trust proxy headers.
	TrustProxy bool
	// Access log file path.
	AccessLog string
}

func ServerFromEnv() Server {
	return Server{
		Debug:      getEnvOrDefault("PRISME_DEBUG", "false") != "false",
		TrustProxy: getEnvOrDefault("PRISME_TRUST_PROXY", "false") != "false",
		AccessLog:  getEnvOrDefault("PRISME_ACCESS_LOG", "/dev/stdout"),
	}
}
