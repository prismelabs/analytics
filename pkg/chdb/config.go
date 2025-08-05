package chdb

import (
	"fmt"
	"os"

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
	finfo, err := os.Stat(c.Path)
	if err == nil {
		if !finfo.IsDir() {
			return fmt.Errorf("chdb path isn't a directory")
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("invalid chdb path: %w", err)
	}

	return nil
}
