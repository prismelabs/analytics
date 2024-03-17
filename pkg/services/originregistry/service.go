package originregistry

import "context"

// Service define an origin registry management service.
type Service interface {
	IsOriginRegistered(context.Context, string) (bool, error)
}
