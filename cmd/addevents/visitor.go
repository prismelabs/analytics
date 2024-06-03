package main

import (
	"time"
)

// Execute routine of a pageview visitor.
func (a App) pageviewVisitor(ch chan<- *Session, entryPageviewTime time.Time) {
	// domain := randomItem(a.cfg.Domains)
	// visitorId := randomVisitorId(a.cfg.VisitorIdsRange)
	// countryCode := randomCountryCode()
	// sessionUuid := uuid.Must(uuid.NewV7())
	// // Update sessionUuid timestamp.
	// copy(sessionUuid[:], binary.BigEndian.AppendUint64(nil, uint64((entryPageviewTime.UTC().Unix()*1000)<<16)))
	// sessionUuid[6] = 0x70 // version byte.
	// entryPath := randomPathName()
	//
	// session := &Session{
	// 	domain:         domain,
	// 	entryPath:      entryPath,
	// 	exitTimestamp:  time.Unix(sessionUuid.Time().UnixTime()),
	// 	exitPath:       entryPath,
	// 	client:         randomDesktopClient(),
	// 	referrerDomain: "direct",
	// 	countryCode:    countryCode,
	// 	visitorId:      visitorId,
	// 	sessionUuid:    sessionUuid,
	// 	utmSource:      "",
	// 	utmMedium:      "",
	// 	utmCampaign:    "",
	// 	utmTerm:        "",
	// 	utmContent:     "",
	// 	pageviews:      1,
	// 	sign:           1,
	// }
	//
	// isExternal := rand.Float64() > a.cfg.DirectTrafficRate
	// if isExternal {
	// 	session.referrerDomain = randomExternalDomain()
	// } else { // direct traffic.
	// 	session.utmSource = randomExternalDomain()
	// 	session.utmMedium = randomItem([]string{"ppc", "foo"})
	// 	session.utmCampaign = randomItem([]string{"spring_sale", "black_friday", "christmas_sale"})
	// 	session.utmTerm = randomItem([]string{"running+shoes", "screen", "computer", "gaming+computer"})
	// 	session.utmContent = randomItem([]string{"logolink", "textlink"})
	// }
	//
	// ch <- session
	// a.metrics.visits.Add(1)
	//
	// if rand.Float64() < a.cfg.BounceRate {
	// 	// Bounce.
	// 	a.metrics.bounces.Add(1)
	// 	return
	// }
	//
	// for {
	// 	// Copy session to avoid data race.
	// 	session := *session
	//
	// 	// Cancel previous session.
	// 	session.sign = -1
	// 	ch <- &session
	//
	// 	// Add new session event.
	// 	session.exitTimestamp = session.exitTimestamp.Add(randomMinute())
	// 	session.exitPath = randomPathName()
	// 	session.sign = 1
	// 	session.pageviews++
	// 	ch <- &session
	//
	// 	if rand.Float64() < a.cfg.ExitRate {
	// 		// Exit.
	// 		return
	// 	}
	// }
}
