package sourceregistry

import "context"

// Source define a unique source identifier. It can be a domain name, a UUID or
// anything that can be represented as a string.
type Source interface {
	SourceString() string
}

// Service define a source registry management service.
type Service interface {
	IsSourceRegistered(context.Context, Source) (bool, error)
}
