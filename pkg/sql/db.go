package sql

import "context"

// DB defines a common interface to interact with databases.
type DB interface {
	Exec(ctx context.Context, query string, args ...any) error
	Query(ctx context.Context, query string, args ...any) (QueryResult, error)
	QueryRow(ctx context.Context, query string, args ...any) Row
}

// QueryResult define result of a query.
type QueryResult interface {
	Next() bool
	Scan(...any) error
	Close() error
}

// Row is the result of calling DB.QueryRow to select a single row.
type Row interface {
	Err() error
	Scan(...any) error
}
