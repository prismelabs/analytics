package eventdb

import (
	"context"
	"fmt"
	"iter"
	"maps"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/prismelabs/analytics/pkg/log"
	"github.com/prismelabs/analytics/pkg/services/teardown"
)

// Service define an event database service.
type Service interface {
	Exec(ctx context.Context, query string, args ...any) error
	Query(ctx context.Context, query string, args ...any) (QueryResult, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
	DriverName() string
	Driver() any
}

// QueryResult define result of a query.
type QueryResult interface {
	Next() bool
	Scan(...any) error
	Close() error
}

// Row is the result of calling Service.QueryRow to select a single row.
type Row interface {
	Err() error
	Scan(...any) error
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
