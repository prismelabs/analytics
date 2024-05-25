package eventstore

import (
	"context"

	"github.com/prismelabs/analytics/pkg/event"
)

// Service define an event storage service.
type Service interface {
	StorePageView(context.Context, *event.PageView) error
	StoreSession(context.Context, *event.Session) error
	StoreCustom(context.Context, *event.Custom) error
}
