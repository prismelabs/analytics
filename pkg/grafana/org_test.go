package grafana

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/prismelabs/prismeanalytics/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestIntegCreateOrganization(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cfg := config.GrafanaFromEnv()
	cli := ProvideClient(cfg)

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

func TestIntegGetOrgByID(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cfg := config.GrafanaFromEnv()
	cli := ProvideClient(cfg)

	t.Run("NonExistentOrganization", func(t *testing.T) {
		orgName, err := cli.GetOrgByID(context.Background(), 9999999)
		require.Error(t, err)
		require.Equal(t, ErrGrafanaOrgNotFound, err)
		require.Equal(t, "", orgName)
	})
	t.Run("ExistentOrganization", func(t *testing.T) {
		expectedOrgName := fmt.Sprintf("foo-%v", rand.Int())
		orgID, err := cli.CreateOrg(context.Background(), expectedOrgName)
		require.NoError(t, err)

		actualOrgName, err := cli.GetOrgByID(context.Background(), orgID)
		require.NoError(t, err)
		require.Equal(t, expectedOrgName, actualOrgName)
	})
}

func TestIntegFindByName(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cfg := config.GrafanaFromEnv()
	cli := ProvideClient(cfg)

	t.Run("NonExistentOrganization", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())

		orgId, err := cli.FindOrgByName(context.Background(), orgName)
		require.Error(t, err)
		require.Equal(t, ErrGrafanaOrgNotFound, err)
		require.Equal(t, OrgId(0), orgId)
	})
	t.Run("ExistentOrganization", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		_, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		orgId, err := cli.FindOrgByName(context.Background(), orgName)
		require.NoError(t, err)
		require.NotEqual(t, OrgId(0), orgId)
	})
}

func TestIntegGetOrCreateOrg(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cfg := config.GrafanaFromEnv()
	cli := ProvideClient(cfg)

	t.Run("NonExistentOrganization", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())

		orgId, err := cli.GetOrCreateOrg(context.Background(), orgName)
		require.NoError(t, err)
		require.NotEqual(t, OrgId(0), orgId)
	})
	t.Run("ExistentOrganization", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())

		expectedOrgId, err := cli.GetOrCreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		actualOrgId, err := cli.GetOrCreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		require.Equal(t, expectedOrgId, actualOrgId)
	})
}

func TestIntegUpdateOrgName(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cfg := config.GrafanaFromEnv()
	cli := ProvideClient(cfg)

	t.Run("NonExistentOrganization", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		err := cli.UpdateOrgName(context.Background(), -1, orgName)
		require.Error(t, err)
		require.Equal(t, `failed to update grafana organization name: 500 {"message":"Failed to update organization","traceID":""}`, err.Error())
	})

	t.Run("ExistentOrganization", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())

		orgId, err := cli.GetOrCreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		newOrgName := fmt.Sprintf("foo-%v", rand.Int())
		err = cli.UpdateOrgName(context.Background(), orgId, newOrgName)
		require.NoError(t, err)

		actualOrgId, err := cli.FindOrgByName(context.Background(), newOrgName)
		require.NoError(t, err)
		require.Equal(t, orgId, actualOrgId)
	})
}

func TestIntegListOrgUsers(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cfg := config.GrafanaFromEnv()
	cli := ProvideClient(cfg)

	t.Run("NonExistentOrganization", func(t *testing.T) {
		users, err := cli.ListOrgUsers(context.Background(), -1)
		require.Error(t, err)
		require.Equal(t, ErrGrafanaOrgNotFound, err)
		require.Nil(t, users)
	})

	t.Run("ExistentOrganization", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.GetOrCreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		users, err := cli.ListOrgUsers(context.Background(), orgId)
		require.NoError(t, err)
		require.Len(t, users, 1)
		require.Equal(t, orgId, users[0].OrgId)
		require.Equal(t, UserId(0), users[0].Id)
		require.Equal(t, RoleAdmin, users[0].Role)
	})
}
