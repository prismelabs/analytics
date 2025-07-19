package grafana

import (
	"flag"
	"io"
	"testing"

	"github.com/negrel/configue"
	"github.com/prismelabs/analytics/pkg/log"
)

func init() {
	testing.Init()
	flag.Parse()
	if testing.Short() {
		return
	}

	logger := log.NewLogger("test_grafana_logger", io.Discard, false)

	var cfg Config
	figue := configue.New(
		"",
		configue.PanicOnError,
		configue.NewEnv("PRISME"),
	)
	cfg.RegisterOptions(figue)
	_ = figue.Parse()

	client := NewClient(cfg)
	WaitHealthy(logger, client, 10)
}
