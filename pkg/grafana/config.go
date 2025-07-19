package grafana

import (
	"errors"
	"net/url"

	"github.com/negrel/configue"
	"github.com/negrel/secrecy"
	"github.com/prismelabs/analytics/pkg/options"
)

// Grafana related options.
type Config struct {
	Url      string
	User     secrecy.SecretString
	Password secrecy.SecretString
	OrgId    int64
}

// RegisterOptions registers Config fields as options.
func (c *Config) RegisterOptions(f *configue.Figue) {
	f.StringVar(&c.Url, "grafana.url", "", "Grafana server `url`")
	f.Var(options.Secret{SecretString: &c.User}, "grafana.user", "Grafana user")
	f.Var(options.Secret{SecretString: &c.Password}, "grafana.password", "Grafana password")
	f.Int64Var(&c.OrgId, "grafana.org.id", 1, "Grafana orgianization ID")
}

// Validate validates configuration options.
func (c *Config) Validate() error {
	var errs []error

	_, err := url.ParseRequestURI(c.Url)
	if err != nil {
		errs = append(errs, errors.New("grafana url must be a valid URL"))
	}
	if c.User.Secret == nil || c.User.ExposeSecret() == "" {
		errs = append(errs,
			errors.New("grafana user missing"))
	}
	if c.Password.Secret == nil || c.Password.ExposeSecret() == "" {
		errs = append(errs, errors.New("grafana password missing"))
	}
	if c.OrgId <= 0 {
		errs = append(errs, errors.New("grafana org id must be greater than 0"))
	}
	return errors.Join(errs...)
}
