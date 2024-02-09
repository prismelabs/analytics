package grafana

import (
	"flag"
	"os"
	"testing"

	"github.com/prismelabs/prismeanalytics/pkg/config"
	"github.com/prismelabs/prismeanalytics/pkg/grafana"
	"github.com/prismelabs/prismeanalytics/pkg/log"
)

func init() {
	testing.Init()
	flag.Parse()
	if testing.Short() {
		return
	}

	logger := log.NewLogger("test_grafana_logger", os.Stderr, false)

	client := grafana.ProvideClient(config.GrafanaFromEnv())
	grafana.WaitHealthy(logger, client, 10)
}
