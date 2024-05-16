package main

import (
	"math/rand"
	"time"
)

// Execute routine of a pageview visitor.
func (a App) pageviewVisitor(ch chan<- *Pageview, entryPageviewTime time.Time) {
	domain := randomItem(a.cfg.Domains)
	os := randomOS()
	browser := randomBrowser()
	device := "benchbot"
	visitorId := randomVisitorId(a.cfg.VisitorIdsRange)
	countryCode := randomCountryCode()
	sessionId := rand.Uint64()

	entryPageview := Pageview{
		timestamp:      entryPageviewTime,
		domain:         domain,
		pathname:       randomPathName(),
		os:             os,
		browser:        browser,
		device:         device,
		referrerDomain: "direct",
		countryCode:    countryCode,
		visitorId:      visitorId,
		sessionId:      sessionId,
		entryTimestamp: entryPageviewTime,
	}

	isExternal := rand.Float64() > a.cfg.DirectTrafficRate
	if isExternal {
		entryPageview.referrerDomain = randomExternalReferrerDomain()
	}

	ch <- &entryPageview
	a.metrics.visits.Add(1)

	if rand.Float64() < a.cfg.BounceRate {
		// Bounce.
		a.metrics.bounces.Add(1)
		return
	}

	for {
		entryPageviewTime = entryPageviewTime.Add(-randomMinute())
		pageview := Pageview{
			timestamp:      entryPageviewTime,
			domain:         domain,
			pathname:       randomPathName(),
			os:             os,
			browser:        browser,
			device:         device,
			referrerDomain: domain, // Internal traffic.
			countryCode:    countryCode,
			visitorId:      visitorId,
			sessionId:      sessionId,
			entryTimestamp: entryPageviewTime,
		}

		ch <- &pageview

		if rand.Float64() < a.cfg.ExitRate {
			// Exit.
			return
		}
	}
}
