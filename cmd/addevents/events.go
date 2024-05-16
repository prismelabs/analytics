package main

import "time"

type Pageview struct {
	datetime       time.Time
	domain         string
	pathname       string
	os             string
	browser        string
	device         string
	referrerDomain string
	countryCode    string
	visitorId      string
	sessionId      uint64
}
