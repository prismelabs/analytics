package grafana

import (
	"flag"
	"io"
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

	logger := log.NewLogger("test_grafana_logger", io.Discard, false)

	client := grafana.ProvideClient(grafana.ProvideConfig(logger))
	grafana.WaitHealthy(logger, client, 10)
}
