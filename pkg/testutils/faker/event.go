//go:build test

package faker

import (
	"math/rand"
	"time"

	"github.com/prismelabs/analytics/pkg/event"
)

// Session returns a random valid event.Session with 0 page view.
func Session() event.Session {
	return event.Session{
		PageUri:       Uri(),
		ReferrerUri:   ReferrerUri(rand.Int()%4 != 0),
		Client:        UapClient(),
		CountryCode:   CountryCode(),
		VisitorId:     "prisme_" + String(AlphaNum, 8),
		SessionUuid:   UuidV7(Time(-6 * time.Hour)),
		Utm:           event.UtmParams{},
		PageviewCount: 0,
	}
}

// PageView returns a random valid event.PageView based on given session.
func PageView(session event.Session) event.PageView {
	return event.PageView{
		Session: session,
		Timestamp: session.SessionTime().Add(
			time.Duration(session.PageviewCount) * time.Minute,
		),
		PageUri: PageUri(session),
		Status:  200,
	}
}

// FileDownload returns a random valid event.FileDownload.
func FileDownload(session event.Session) event.FileDownload {
	return event.FileDownload{
		Timestamp: session.SessionTime().Add(
			time.Duration(session.PageviewCount) * time.Minute,
		),
		PageUri: PageUri(session),
		Session: session,
		FileUrl: PageUri(session),
	}
}

// OutboundLinkClick returns a random valid event.OutboundLinkClick.
func OutboundLinkClick(session event.Session) event.OutboundLinkClick {
	return event.OutboundLinkClick{
		Timestamp: session.SessionTime().Add(
			time.Duration(session.PageviewCount) * time.Minute,
		),
		PageUri: PageUri(session),
		Session: session,
		Link:    Uri(),
	}
}

// CustomEvent returns a random event.Custom.
func CustomEvent(session event.Session) event.Custom {
	return event.Custom{
		Timestamp: session.SessionTime().Add(
			time.Duration(session.PageviewCount) * time.Minute,
		),
		PageUri: PageUri(session),
		Session: session,
		Name:    "click",
		Keys:    []string{"x", "y"},
		Values:  []string{"100", "200"},
	}
}
