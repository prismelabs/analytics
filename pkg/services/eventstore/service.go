package eventstore

import (
	"context"

	"github.com/prismelabs/analytics/pkg/event"
)

// Service define an event storage service.
type Service interface {
	StorePageViewEvent(context.Context, event.PageView) error
}
