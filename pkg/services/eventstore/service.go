package eventstore

import (
	"context"

	"github.com/prismelabs/analytics/pkg/event"
)

// Service define an event storage service.
type Service interface {
	StorePageView(context.Context, *event.PageView) error
	StoreCustom(context.Context, *event.Custom) error
}
