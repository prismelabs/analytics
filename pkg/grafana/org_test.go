package grafana

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestIntegCreateOrganization(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	cli := NewClient(cfg)

	t.Run("NonExistentOrganization", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)
		require.NotEqual(t, OrgId(0), orgId)
	})
	t.Run("ExistentOrganization", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		{
			orgId, err := cli.CreateOrg(context.Background(), orgName)
			require.NoError(t, err)
			require.NotEqual(t, OrgId(0), orgId)
		}

		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.Error(t, err)
		require.Equal(t, ErrGrafanaOrgAlreadyExists, err)
		require.Equal(t, OrgId(0), orgId)
	})
}
