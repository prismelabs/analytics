package full

import (
	"context"
	"time"

	grafanaCli "github.com/prismelabs/prismeanalytics/pkg/grafana"
	"github.com/prismelabs/prismeanalytics/pkg/log"
	"github.com/prismelabs/prismeanalytics/pkg/services/grafana"
	"github.com/prismelabs/prismeanalytics/pkg/wired"
)

// ProvideSetup is a wire provider that performs setup of full server.
func ProvideSetup(logger log.Logger, cli grafanaCli.Client, grafanaService grafana.Service) wired.Setup {
	grafanaCli.WaitHealthy(logger, cli, 5)
	logger.Info().Msg("setting up grafana datasource and dashboards...")
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := grafanaService.SetupDatasourceAndDashboards(ctx, 1)
		if err != nil {
			logger.Panic().Err(err).Msg("failed to setup grafana datasource and dashboards")
		}
	}
	logger.Info().Msg("grafana datasource and dashboards successfully configured.")

	return wired.Setup{}
}
