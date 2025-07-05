package grafana

import (
	"flag"
	"io"
	"testing"

	"github.com/prismelabs/analytics/pkg/log"
)

func init() {
	testing.Init()
	flag.Parse()
	if testing.Short() {
		return
	}

	logger := log.NewLogger("test_grafana_logger", io.Discard, false)

	client := ProvideClient(configFromEnv())
	WaitHealthy(logger, client, 10)
}
