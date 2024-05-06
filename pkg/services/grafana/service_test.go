package grafana

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/prismelabs/analytics/pkg/config"
	"github.com/prismelabs/analytics/pkg/grafana"
	"github.com/stretchr/testify/require"
)

func TestIntegService(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cli := grafana.ProvideClient(config.GrafanaFromEnv())
	service := ProvideService(cli, config.ClickhouseFromEnv())
	ctx := context.Background()

	t.Run("SetupDatasourceAndDashboards", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(ctx, orgName)
		require.NoError(t, err)

		err = service.SetupDatasourceAndDashboards(context.Background(), orgId)
		require.NoError(t, err)

		// Check folder was created.
		folders, err := cli.ListFolders(ctx, orgId, 100, 0)
		require.NoError(t, err)
		require.Len(t, folders, 1)
		require.Equal(t, folders[0].Title, "Prisme Analytics")

		// Check dashboards were created.
		dashboards, err := cli.SearchDashboards(ctx, orgId, 100, 0)
		require.NoError(t, err)
		require.Len(t, dashboards, 1)
		require.Equal(t, dashboards[0].Title, "Web Analytics")
	})
}
