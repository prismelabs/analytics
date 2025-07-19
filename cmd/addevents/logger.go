package main

import (
	"os"

	"github.com/prismelabs/analytics/pkg/log"
)

// NewLogger returns a new configured logger.
func NewLogger() log.Logger {
	logger := log.New("app", os.Stderr, true)
	err := logger.TestOutput()
	if err != nil {
		panic(err)
	}

	return logger
}
