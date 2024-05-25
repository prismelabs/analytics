package main

import (
	"math/big"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// Execute routine of a pageview visitor.
func (a App) pageviewVisitor(ch chan<- any, entryPageviewTime time.Time) {
	domain := randomItem(a.cfg.Domains)
	visitorId := randomVisitorId(a.cfg.VisitorIdsRange)

	session := Session{
		timestamp:      entryPageviewTime,
		domain:         domain,
		pathname:       randomPathName(),
		os:             randomOS(),
		browser:        randomBrowser(),
		device:         "benchbot",
		referrerDomain: "direct",
		countryCode:    randomCountryCode(),
		visitorId:      visitorId,
		sessionId:      uuid.Must(uuid.NewV7()),
	}

	isExternal := rand.Float64() > a.cfg.DirectTrafficRate
	if isExternal {
		session.referrerDomain = randomExternalReferrerDomain()
	}

	ch <- &session
	a.metrics.sessions.Add(1)

	if rand.Float64() < a.cfg.BounceRate {
		// Bounce.
		a.metrics.bounces.Add(1)
		return
	}

	for {
		entryPageviewTime = entryPageviewTime.Add(-randomMinute())
		pageview := Pageview{
			timestamp: entryPageviewTime,
			domain:    domain,
			pathname:  randomPathName(),
			visitorId: visitorId,
			sessionId: big.NewInt(0).SetBytes(session.sessionId[:]),
		}

		ch <- &pageview

		if rand.Float64() < a.cfg.ExitRate {
			// Exit.
			return
		}
	}
}
