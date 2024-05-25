package main

import (
	"math/big"
	"time"

	"github.com/google/uuid"
)

type Pageview struct {
	timestamp time.Time
	domain    string
	pathname  string
	visitorId string
	sessionId *big.Int
}

type Session struct {
	timestamp      time.Time
	domain         string
	pathname       string
	os             string
	browser        string
	device         string
	referrerDomain string
	countryCode    string
	visitorId      string
	sessionId      uuid.UUID
}
