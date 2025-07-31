package eventdb

import (
	"fmt"
	"slices"
	"strings"

	"github.com/negrel/configue"
)

// Config of eventdb service.
type Config struct {
	Driver string
}

// RegisterOptions registers Config fields as options.
func (c *Config) RegisterOptions(f *configue.Figue) {
	drivers := "(" + strings.Join(slices.Collect(Drivers()), ", ") + ")"
	f.StringVar(&c.Driver, "eventdb.driver", "clickhouse", "event database `driver` to use "+drivers)
}

// Validate validates configuration options.
func (c *Config) Validate() error {
	_, ok := dbFactory[c.Driver]
	if ok {
		return nil
	}

	return fmt.Errorf("unsupported event database driver %s", c.Driver)
}
