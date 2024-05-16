package main

import (
	"math/rand"
	"time"
)

// Execute routine of a pageview visitor.
func (a App) pageviewVisitor(ch chan<- *Pageview, pageviewTime time.Time) {
	domain := randomItem(a.cfg.Domains)
	os := randomOS()
	browser := randomBrowser()
	device := "benchbot"
	visitorId := randomVisitorId(a.cfg.VisitorIdsRange)
	countryCode := randomCountryCode()
	sessionId := rand.Uint64()

	entryPageview := Pageview{
		datetime:       pageviewTime,
		domain:         domain,
		pathname:       randomPathName(),
		os:             os,
		browser:        browser,
		device:         device,
		referrerDomain: "direct",
		countryCode:    countryCode,
		visitorId:      visitorId,
		sessionId:      sessionId,
	}

	isExternal := randomFactors(a.cfg.DirectTrafficFactor, a.cfg.ExternalTrafficFactor) == 1
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
		pageviewTime = pageviewTime.Add(-randomMinute())
		pageview := Pageview{
			datetime:       pageviewTime,
			domain:         domain,
			pathname:       randomPathName(),
			os:             os,
			browser:        browser,
			device:         device,
			referrerDomain: domain, // Internal traffic.
			countryCode:    countryCode,
			visitorId:      visitorId,
			sessionId:      sessionId,
		}

		ch <- &pageview

		if rand.Float64() < a.cfg.ExitRate {
			// Exit.
			return
		}
	}
}
