package main

import "time"

type Pageview struct {
	timestamp      time.Time
	domain         string
	pathname       string
	os             string
	browser        string
	device         string
	referrerDomain string
	countryCode    string
	visitorId      string
	sessionId      uint64
	entryTimestamp time.Time
}
