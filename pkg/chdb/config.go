package chdb

import (
	"github.com/negrel/configue"
)

// chdb options.
type Config struct {
	Path string
}

// RegisterOptions registers Config fields as options.
func (c *Config) RegisterOptions(f *configue.Figue) {
	f.StringVar(&c.Path, "chdb.path", "", "chdb directory `filepath`")
}

// Validate validates configuration options.
func (c *Config) Validate() error {
	return nil
}
