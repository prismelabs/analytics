package main

import (
	"encoding/binary"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/services/uaparser"
)

// Session define a row in prisme.sessions table.
type Session struct {
	domain         string
	entryPath      string
	exitTimestamp  time.Time
	exitPath       string
	client         uaparser.Client
	referrerDomain string
	countryCode    string
	visitorId      string
	sessionUuid    uuid.UUID
	utmSource      string
	utmMedium      string
	utmCampaign    string
	utmTerm        string
	utmContent     string
	pageviews      uint16
	sign           int8
}

// SessionTimestamp returns session creation datetime.
func (s *Session) SessionTimestamp() time.Time {
	return time.Unix(s.sessionUuid.Time().UnixTime()).UTC()
}

// Row encode session as a flat []any slice ready to be sent to ClickHouse.
func (s *Session) Row() (record []any) {
	// Order must match sessions table column order.
	record = append(record,
		s.domain,
		s.entryPath,
		s.exitTimestamp,
		s.exitPath,
		s.visitorId,
		s.sessionUuid,
		s.client.OperatingSystem,
		s.client.BrowserFamily,
		s.client.Device,
		s.referrerDomain,
		s.countryCode,
		s.utmSource,
		s.utmMedium,
		s.utmCampaign,
		s.utmTerm,
		s.utmContent,
		s.pageviews,
		s.sign,
	)

	return
}

// emulateSession emulates a single sessions and returns the number of events generated.
func emulateSession(entryTime time.Time, cfg Config, rowsChan chan<- any) uint64 {
	var eventsCount uint64 = 1

	entryTime = entryTime.Add(randomMinute()).UTC()
	domain := randomItem(cfg.Domains)
	visitorId := randomVisitorId(cfg.VisitorIdsRange)
	countryCode := randomCountryCode()
	sessionUuid := uuid.Must(uuid.NewV7())
	// Update sessionUuid timestamp.
	copy(sessionUuid[:], binary.BigEndian.AppendUint64(nil, uint64((entryTime.Unix()*1000)<<16)))
	sessionUuid[6] = 0x70 // version byte.
	entryPath := randomPathName()

	session := Session{
		domain:         domain,
		entryPath:      entryPath,
		exitTimestamp:  entryTime,
		exitPath:       entryPath,
		client:         randomDesktopClient(),
		referrerDomain: "direct",
		countryCode:    countryCode,
		visitorId:      visitorId,
		sessionUuid:    sessionUuid,
		utmSource:      "",
		utmMedium:      "",
		utmCampaign:    "",
		utmTerm:        "",
		utmContent:     "",
		pageviews:      1,
		sign:           1,
	}

	if rand.Float64() < cfg.MobileRate {
		session.client = randomMobileClient()
	}

	isExternal := rand.Float64() > cfg.DirectTrafficRate
	if isExternal {
		session.referrerDomain = randomExternalDomain()
	} else { // direct traffic.
		session.utmSource = randomExternalDomain()
		session.utmMedium = randomItem([]string{"ppc", "foo"})
		session.utmCampaign = randomItem([]string{"spring_sale", "black_friday", "christmas_sale"})
		session.utmTerm = randomItem([]string{"running+shoes", "screen", "computer", "gaming+computer"})
		session.utmContent = randomItem([]string{"logolink", "textlink"})
	}

	rowsChan <- session

	for rand.Float64() < cfg.CustomEventsRate {
		rowsChan <- randomCustomEvent(session)
		eventsCount++
	}

	if rand.Float64() < cfg.BounceRate {
		// Bounce.
		return 1
	}

	for {
		// Cancel previous session.
		session.sign = -1
		rowsChan <- session

		// Add new session event.
		session.exitTimestamp = session.exitTimestamp.Add(randomMinute())
		session.exitPath = randomPathName()
		session.sign = 1
		session.pageviews++
		rowsChan <- session
		eventsCount++

		for rand.Float64() < cfg.CustomEventsRate {
			rowsChan <- randomCustomEvent(session)
			eventsCount++
		}

		if rand.Float64() < cfg.ExitRate {
			// Exit.
			return eventsCount
		}
	}
}
