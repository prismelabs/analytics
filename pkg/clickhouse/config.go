package clickhouse

import (
	"errors"
	"net"

	"github.com/negrel/configue"
	"github.com/negrel/secrecy"
	"github.com/prismelabs/analytics/pkg/options"
)

// ClickHouse connection options.
type Config struct {
	TlsEnabled bool
	HostPort   string
	Database   string
	User       secrecy.SecretString
	Password   secrecy.SecretString
}

// RegisterOptions registers Config fields as options.
func (c *Config) RegisterOptions(f *configue.Figue) {
	f.BoolVar(&c.TlsEnabled, "clickhouse.tls", false, "use a TLS connection for ClickHouse")
	f.StringVar(&c.HostPort, "clickhouse.hostport", "", "use `host:port` to connect to ClickHouse")
	f.StringVar(&c.Database, "clickhouse.database", "prisme", "ClickHouse database to use")
	f.Var(options.Secret{SecretString: &c.User}, "clickhouse.user", "ClickHouse user")
	f.Var(options.Secret{SecretString: &c.Password}, "clickhouse.password", "ClickHouse password")
}

// Validate validates configuration options.
func (c *Config) Validate() error {
	_, _, err := net.SplitHostPort(c.HostPort)
	if err != nil {
		return errors.New("invalid ClickHouse hostport")
	}
	if c.User.Secret == nil || c.User.ExposeSecret() == "" {
		return errors.New("ClickHouse user missing")
	}
	if c.User.Secret == nil || c.Password.ExposeSecret() == "" {
		return errors.New("ClickHouse password missing")
	}

	return nil
}
