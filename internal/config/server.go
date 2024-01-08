package config

// Server specific options.
type Server struct {
	// Sets log level to debug.
	Debug bool
	// Trust proxy headers.
	TrustProxy bool
}

func ServerFromEnv() Server {
	return Server{
		Debug:      getEnvOrDefault("PRISME_DEBUG", "false") != "false",
		TrustProxy: getEnvOrDefault("PRISME_TRUST_PROXY", "false") != "false",
	}
}
