package grafana

import (
	"flag"
	"os"
	"testing"

	"github.com/prismelabs/analytics/pkg/grafana"
	"github.com/prismelabs/analytics/pkg/log"
)

func init() {
	testing.Init()
	flag.Parse()
	if testing.Short() {
		return
	}

	logger := log.NewLogger("test_grafana_logger", os.Stderr, false)

	client := grafana.ProvideClient(grafana.ProvideConfig(logger))
	grafana.WaitHealthy(logger, client, 10)
}
