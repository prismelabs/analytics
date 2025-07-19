package grafana

import (
	"flag"
	"io"
	"testing"

	"github.com/negrel/configue"
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

	figue := configue.New(
		"",
		configue.PanicOnError,
		configue.NewEnv("PRISME"),
	)
	var grafanaCfg grafana.Config
	grafanaCfg.RegisterOptions(figue)
	_ = figue.Parse()

	client := grafana.ProvideClient(grafanaCfg)
	grafana.WaitHealthy(logger, client, 10)
}
