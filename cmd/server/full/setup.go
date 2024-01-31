package full

import (
	"context"
	"time"

	"github.com/prismelabs/prismeanalytics/cmd/server/wired"
	grafanaCli "github.com/prismelabs/prismeanalytics/internal/grafana"
	"github.com/prismelabs/prismeanalytics/internal/log"
	"github.com/prismelabs/prismeanalytics/internal/services/grafana"
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
