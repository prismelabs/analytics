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
	// X-Forwarded-For proxy header.
	ProxyHeader string
}

// ServerFromEnv loads server related options from environment variables.
func ServerFromEnv() Server {
	return Server{
		AccessLog:   GetEnvOrDefault("PRISME_ACCESS_LOG", "/dev/stdout"),
		Debug:       GetEnvOrDefault("PRISME_DEBUG", "false") != "false",
		Port:        uint16(ParseUintEnvOrDefault("PRISME_PORT", 80, 16)),
		TrustProxy:  GetEnvOrDefault("PRISME_TRUST_PROXY", "false") != "false",
		ProxyHeader: GetEnvOrDefault("PRISME_PROXY_HEADER", "X-Forwarded-For"),
	}
}
