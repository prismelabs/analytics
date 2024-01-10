package config

// Server specific options.
type Server struct {
	// Access log file path.
	AccessLog string
	// Sets log level to debug.
	Debug bool
	// Listening port.
	Port uint16
	// Trust proxy headers.
	TrustProxy bool
}

// ServerFromEnv loads server related options from environment variables.
func ServerFromEnv() Server {
	return Server{
		AccessLog:  getEnvOrDefault("PRISME_ACCESS_LOG", "/dev/stdout"),
		Debug:      getEnvOrDefault("PRISME_DEBUG", "false") != "false",
		Port:       uint16(parseUintEnvOrDefault("PRISME_PORT", 80, 16)),
		TrustProxy: getEnvOrDefault("PRISME_TRUST_PROXY", "false") != "false",
	}
}
