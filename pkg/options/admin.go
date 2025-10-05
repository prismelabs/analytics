package options

import (
	"net"

	"github.com/negrel/configue"
)

// Admin HTTP server options.
type Admin struct {
	// host:port address of admin http server.
	HostPort string
}

// RegisterOptions registers options in provided Figue.
func (a *Admin) RegisterOptions(f *configue.Figue) {
	f.StringVar(&a.HostPort, "admin.hostport", "127.0.0.1:9090", "use `host:port` for administration HTTP server")
}

// Validate validates configuration options.
func (a *Admin) Validate() error {
	_, _, err := net.SplitHostPort(a.HostPort)
	if err != nil {
		return err
	}
	return nil
}
