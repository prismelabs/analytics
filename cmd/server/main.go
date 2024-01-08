package main

import (
	"os"

	"github.com/prismelabs/prismeanalytics/internal/log"
)

func main() {
	logger := log.NewLogger("app", os.Stderr, false)
	log.TestLoggers(logger)
}
