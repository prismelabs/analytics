package eventdb

import (
	"fmt"
	"iter"
	"maps"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/teardown"
	"github.com/prismelabs/analytics/pkg/sql"
)

// Service define an event database service.
type Service interface {
	sql.DB
	DriverName() string
	Driver() any
}

var dbFactory = map[string]func(
	logger log.Logger,
	cfg any,
	source source.Driver,
	teardown teardown.Service,
) (Service, error){}

// NewService returns a new eventdb service.
func NewService(
	cfg Config,
	driverCfg any,
	logger log.Logger,
	source source.Driver,
	teardown teardown.Service,
) (Service, error) {
	fact, ok := dbFactory[cfg.Driver]
	if !ok {
		return nil, fmt.Errorf("unknown eventdb driver: %v", cfg.Driver)
	}
	return fact(logger, driverCfg, source, teardown)
}

// Drivers returns an iterator over available drivers.
func Drivers() iter.Seq[string] {
	return maps.Keys(dbFactory)
}
