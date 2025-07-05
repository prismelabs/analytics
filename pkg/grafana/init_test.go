package grafana

import (
	"flag"
	"os"
	"testing"

	"github.com/prismelabs/analytics/pkg/log"
)

func init() {
	testing.Init()
	flag.Parse()
	if testing.Short() {
		return
	}

	logger := log.NewLogger("test_grafana_logger", os.Stderr, false)

	client := ProvideClient(configFromEnv())
	WaitHealthy(logger, client, 10)
}
