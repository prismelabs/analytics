package main

import (
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2/utils"
)

var OSs = []string{
	"Windows",
	"Linux",
	"Mac OS X",
	"iOS",
	"Android",
}

func randomOS() string {
	return randomItem(OSs)
}

var browsers = []string{
	"Firefox",
	"Chrome",
	"Edge",
	"Opera",
}

func randomBrowser() string {
	return randomItem(browsers)
}

var pathnames = []string{
	"/foo/bar/qux",
	"/foo/bar/",
	"/foo/bar",
	"/foo",
	"/blog",
	"/blog/misc/a-nice-post",
	"/blog/misc/another-nice-post",
	"/contact",
	"/terms-of-service",
	"/privacy",
}

func randomPathName() string {
	return randomItem(pathnames)
}

var referrerDomains = []string{
	"twitter.com",
	"facebook.com",
	"google.com",
	"direct",
}

func randomReferrerDomain(extraDomains []string) string {
	if rand.Int()%2 == 0 {
		return randomItem(extraDomains)
	}

	return randomItem(referrerDomains)
}

func randomMinute() time.Duration {
	return time.Duration((rand.Int() % 60)) * time.Minute
}

func randomItem[T any](slice []T) T {
	index := rand.Int() % len(slice)
	return slice[index]
}

const (
	alphaLower = "abcdefghijklmnopqrstuvwxyz"
	alphaUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alpha      = alphaLower + alphaUpper
	num        = "0123456789"
	alphaNum   = alpha + num
)

func randomString(charset string, length int) string {
	buf := make([]byte, length)

	for i := 0; i < length; i++ {
		buf[i] = charset[rand.Intn(len(charset)-1)]
	}

	return utils.UnsafeString(buf)
}
