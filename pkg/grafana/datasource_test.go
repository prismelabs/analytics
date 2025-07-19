package grafana

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestIntegCreateDatasource(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	cli := ProvideClient(cfg)

	t.Run("ExistentDatasourceType", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		dsName := fmt.Sprintf("datasource-%v", rand.Int())

		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		ds, err := cli.CreateDatasource(context.Background(), orgId, dsName, "grafana-clickhouse-datasource", false)
		require.NoError(t, err)

		require.NotEqual(t, Uid{}, ds.Uid)
		ds.Uid = Uid{}

		require.NotEqual(t, int64(0), ds.Id)
		ds.Id = 0

		require.Equal(t, Datasource{
			Access:      "proxy",
			BasicAuth:   false,
			Database:    "",
			Id:          0,
			IsDefault:   true,
			JSONData:    map[string]any{},
			Name:        dsName,
			OrgId:       orgId,
			ReadOnly:    false,
			Type:        "grafana-clickhouse-datasource",
			TypeLogoUrl: "public/plugins/grafana-clickhouse-datasource/img/logo.svg",
			TypeName:    "",
			Uid:         Uid{},
			URL:         "",
			User:        "",
			Version:     1,
		}, ds)
	})

	t.Run("NonExistentDatasourceType", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		dsName := fmt.Sprintf("datasource-%v", rand.Int())

		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		ds, err := cli.CreateDatasource(context.Background(), orgId, dsName, "non-existent-datasource", false)
		require.NoError(t, err)

		require.NotEqual(t, Uid{}, ds.Uid)
		ds.Uid = Uid{}

		require.NotEqual(t, int64(0), ds.Id)
		ds.Id = 0

		require.Equal(t, Datasource{
			Access:      "proxy",
			BasicAuth:   false,
			Database:    "",
			Id:          0,
			IsDefault:   true,
			JSONData:    map[string]any{},
			Name:        dsName,
			OrgId:       orgId,
			ReadOnly:    false,
			Type:        "non-existent-datasource",
			TypeLogoUrl: "public/img/icn-datasource.svg",
			TypeName:    "",
			Uid:         Uid{},
			URL:         "",
			User:        "",
			Version:     1,
		}, ds)
	})

	t.Run("NameAlreadyTaken", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		dsName := fmt.Sprintf("datasource-%v", rand.Int())

		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		_, err = cli.CreateDatasource(context.Background(), orgId, dsName, "grafana-clickhouse-datasource", false)
		require.NoError(t, err)

		_, err = cli.CreateDatasource(context.Background(), orgId, dsName, "grafana-clickhouse-datasource", false)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrGrafanaDatasourceAlreadyExists)
	})
}

func TestIntegListDatasourcesForCurrentOrg(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	cli := ProvideClient(cfg)

	t.Run("NoDatasources", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())

		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		datasources, err := cli.ListDatasources(context.Background(), orgId)
		require.NoError(t, err)
		require.Len(t, datasources, 0)
	})
	t.Run("SingleDatasource", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		dsName := fmt.Sprintf("datasource-%v", rand.Int())

		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		ds, err := cli.CreateDatasource(context.Background(), orgId, dsName, "grafana-clickhouse-datasource", false)
		require.NoError(t, err)

		datasources, err := cli.ListDatasources(context.Background(), orgId)
		require.NoError(t, err)
		require.Len(t, datasources, 1)
		require.Equal(t, ds.Uid.String(), datasources[0].Uid.String())
	})
	t.Run("MultipleDatasources", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())

		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		var expectedDatasources []Datasource
		for i := 0; i < 10; i++ {
			dsName := fmt.Sprintf("datasource-%v", rand.Int())
			ds, err := cli.CreateDatasource(context.Background(), orgId, dsName, "grafana-clickhouse-datasource", false)
			require.NoError(t, err)

			expectedDatasources = append(expectedDatasources, ds)
		}

		datasources, err := cli.ListDatasources(context.Background(), orgId)
		require.NoError(t, err)
		require.Len(t, datasources, len(expectedDatasources))

		for _, expected := range expectedDatasources {
			found := false
			for _, actual := range datasources {
				if actual.Uid.String() == expected.Uid.String() {
					found = true
				}
			}

			require.Truef(t, found, "failed to found %v in list %v", expected, datasources)
		}
	})
}

func TestIntegUpdateDatasource(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	cli := ProvideClient(cfg)

	t.Run("NonExistentDatasource", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())

		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		err = cli.UpdateDatasource(context.Background(), orgId, Datasource{
			Access:      "",
			BasicAuth:   false,
			Database:    "",
			Id:          0,
			IsDefault:   false,
			JSONData:    map[string]any{},
			Name:        "",
			OrgId:       0,
			ReadOnly:    false,
			Type:        "",
			TypeLogoUrl: "",
			TypeName:    "",
			Uid:         Uid{},
			URL:         "",
			User:        "",
			Version:     0,
		})
		require.Error(t, err)
		require.Equal(t, `failed to update grafana datasource: 400 {"message":"bad request data"}`, err.Error())
	})

	t.Run("ExistentDatasource", func(t *testing.T) {
		t.Run("ChangeName", func(t *testing.T) {
			orgName := fmt.Sprintf("foo-%v", rand.Int())
			dsName := fmt.Sprintf("datasource-%v", rand.Int())

			orgId, err := cli.CreateOrg(context.Background(), orgName)
			require.NoError(t, err)

			ds, err := cli.CreateDatasource(context.Background(), orgId, dsName, "grafana-clickhouse-datasource", false)
			require.NoError(t, err)

			ds.Name = "custom-name"
			err = cli.UpdateDatasource(context.Background(), orgId, ds)
			require.NoError(t, err)

			datasources, err := cli.ListDatasources(context.Background(), orgId)
			require.NoError(t, err)
			require.Len(t, datasources, 1)
			require.Equal(t, "custom-name", datasources[0].Name)
			require.Equal(t, ds.Uid.String(), datasources[0].Uid.String())
		})

		t.Run("ChangeType", func(t *testing.T) {
			orgName := fmt.Sprintf("foo-%v", rand.Int())
			dsName := fmt.Sprintf("datasource-%v", rand.Int())

			orgId, err := cli.CreateOrg(context.Background(), orgName)
			require.NoError(t, err)

			ds, err := cli.CreateDatasource(context.Background(), orgId, dsName, "grafana-clickhouse-datasource", false)
			require.NoError(t, err)

			ds.Type = "non-existent-datasource"
			err = cli.UpdateDatasource(context.Background(), orgId, ds)
			require.NoError(t, err)

			datasources, err := cli.ListDatasources(context.Background(), orgId)
			require.NoError(t, err)
			require.Len(t, datasources, 1)
			require.Equal(t, "non-existent-datasource", datasources[0].Type)
			require.Equal(t, ds.Uid.String(), datasources[0].Uid.String())
		})

		t.Run("OlderVersion", func(t *testing.T) {
			orgName := fmt.Sprintf("foo-%v", rand.Int())
			dsName := fmt.Sprintf("datasource-%v", rand.Int())

			orgId, err := cli.CreateOrg(context.Background(), orgName)
			require.NoError(t, err)

			ds, err := cli.CreateDatasource(context.Background(), orgId, dsName, "grafana-clickhouse-datasource", false)
			require.NoError(t, err)

			ds.Name = "custom-name"
			err = cli.UpdateDatasource(context.Background(), orgId, ds)
			require.NoError(t, err)

			err = cli.UpdateDatasource(context.Background(), orgId, ds)
			require.Error(t, err)
			require.ErrorIs(t, err, ErrGrafanaDatasourceAlreadyUpdated)
		})
	})
}

func TestIntegDeleteDatasource(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	cli := ProvideClient(cfg)

	t.Run("NonExistentDatasource", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		dsName := fmt.Sprintf("datasource-%v", rand.Int())

		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		err = cli.DeleteDatasourceByName(context.Background(), orgId, dsName)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrGrafanaDatasourceNotFound)
	})

	t.Run("ExistentDatasource", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		dsName := fmt.Sprintf("datasource-%v", rand.Int())

		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		_, err = cli.CreateDatasource(context.Background(), orgId, dsName, "grafana-clickhouse-datasource", false)
		require.NoError(t, err)

		err = cli.DeleteDatasourceByName(context.Background(), orgId, dsName)
		require.NoError(t, err)
	})
}

func TestIntegGetDatasourceByName(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	cli := ProvideClient(cfg)

	t.Run("NonExistentDatasource", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		dsName := fmt.Sprintf("datasource-%v", rand.Int())

		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		ds, err := cli.GetDatasourceByName(context.Background(), orgId, dsName)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrGrafanaDatasourceNotFound)
		require.Equal(t, Datasource{}, ds)
	})

	t.Run("ExistentDatasource", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		dsName := fmt.Sprintf("datasource-%v", rand.Int())

		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		expectedDs, err := cli.CreateDatasource(context.Background(), orgId, dsName, "grafana-clickhouse-datasource", false)
		require.NoError(t, err)

		ds, err := cli.GetDatasourceByName(context.Background(), orgId, dsName)
		require.NoError(t, err)
		require.Equal(t, expectedDs, ds)
	})
}
