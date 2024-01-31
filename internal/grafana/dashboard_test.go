package grafana

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prismelabs/prismeanalytics/internal/config"
	"github.com/stretchr/testify/require"
)

func TestIntegCreateUpdateDashboard(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cfg := config.GrafanaFromEnv()
	cli := ProvideClient(cfg)

	t.Run("EmptyTitle", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		dashboardId, err := cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{}, false)
		require.Error(t, err)
		require.Regexp(t, "Dashboard title cannot be empty", err.Error())
		require.Equal(t, DashboardID{}, dashboardId)
	})

	t.Run("NonExistentTitle", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		dashboardId, err := cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{"title": "Dashboard 1"}, false)
		require.NoError(t, err)
		require.NotEqual(t, DashboardID{}, dashboardId)
	})

	t.Run("AlreadyExistentTitle", func(t *testing.T) {
		t.Run("MissingVersionField", func(t *testing.T) {
			orgName := fmt.Sprintf("foo-%v", rand.Int())
			orgId, err := cli.CreateOrg(context.Background(), orgName)
			require.NoError(t, err)

			dashboardId, err := cli.CreateUpdateDashboard(
				context.Background(), orgId, FolderID{},
				map[string]any{"title": "Dashboard 1"},
				false,
			)
			require.NoError(t, err)
			require.NotEqual(t, DashboardID{}, dashboardId)

			// Update.
			updateDashboardId, err := cli.CreateUpdateDashboard(
				context.Background(), orgId, FolderID{},
				map[string]any{"title": "Dashboard 1 v2", "uid": dashboardId.String()},
				false,
			)
			require.Error(t, err)
			require.Regexp(t, "The dashboard has been changed by someone else", err.Error())
			require.Equal(t, DashboardID{}, updateDashboardId)
		})

		t.Run("WithCorrectVersionField", func(t *testing.T) {
			orgName := fmt.Sprintf("foo-%v", rand.Int())
			orgId, err := cli.CreateOrg(context.Background(), orgName)
			require.NoError(t, err)

			dashboardId, err := cli.CreateUpdateDashboard(
				context.Background(), orgId, FolderID{},
				map[string]any{"title": "Dashboard 1", "version": 1},
				false,
			)
			require.NoError(t, err)
			require.NotEqual(t, DashboardID{}, dashboardId)

			// Get version.
			dashboard, err := cli.GetDashboardByUID(context.Background(), orgId, dashboardId)
			require.NoError(t, err)

			// Update.
			updateDashboardId, err := cli.CreateUpdateDashboard(
				context.Background(), orgId, FolderID{},
				map[string]any{"title": "Dashboard 1 v2", "uid": dashboardId.String(), "version": dashboard.Metadata.Version},
				false,
			)
			require.NoError(t, err)
			require.Equal(t, dashboardId, updateDashboardId)

			t.Run("SecondUpdate/SameVersion", func(t *testing.T) {
				// Update again.
				updateDashboardId, err := cli.CreateUpdateDashboard(
					context.Background(), orgId, FolderID{},
					map[string]any{"title": "Dashboard 1 v2", "uid": dashboardId.String(), "version": dashboard.Metadata.Version},
					false,
				)
				require.Error(t, err)
				require.Regexp(t, "The dashboard has been changed by someone else", err.Error())
				require.Equal(t, DashboardID{}, updateDashboardId)
			})

			t.Run("SecondUpdate/IncrementVersion", func(t *testing.T) {
				// Update again.
				updateDashboardId, err := cli.CreateUpdateDashboard(
					context.Background(), orgId, FolderID{},
					map[string]any{"title": "Dashboard 1 v2", "uid": dashboardId.String(), "version": dashboard.Metadata.Version + 1},
					false,
				)
				require.NoError(t, err)
				require.Equal(t, dashboardId, updateDashboardId)
			})
		})

		t.Run("WithIncorrectVersionField", func(t *testing.T) {
			orgName := fmt.Sprintf("foo-%v", rand.Int())
			orgId, err := cli.CreateOrg(context.Background(), orgName)
			require.NoError(t, err)

			dashboardId, err := cli.CreateUpdateDashboard(
				context.Background(), orgId, FolderID{},
				map[string]any{"title": "Dashboard 1", "version": 1},
				false,
			)
			require.NoError(t, err)
			require.NotEqual(t, DashboardID{}, dashboardId)

			// Update.
			updateDashboardId, err := cli.CreateUpdateDashboard(
				context.Background(), orgId, FolderID{},
				map[string]any{"title": "Dashboard 1 v2", "uid": dashboardId.String(), "version": 10},
				false,
			)
			require.Error(t, err)
			require.Regexp(t, "The dashboard has been changed by someone else", err.Error())
			require.Equal(t, DashboardID{}, updateDashboardId)
		})
	})

	t.Run("Overwrite", func(t *testing.T) {
		t.Run("NonExistentDashboard", func(t *testing.T) {
			orgName := fmt.Sprintf("foo-%v", rand.Int())
			orgId, err := cli.CreateOrg(context.Background(), orgName)
			require.NoError(t, err)

			dashboardId, err := cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{"title": "Dashboard 1"}, true)
			require.NoError(t, err)
			require.NotEqual(t, DashboardID{}, dashboardId)
		})

		t.Run("ExistentDashboard/WithoutVersion/WithUID", func(t *testing.T) {
			orgName := fmt.Sprintf("foo-%v", rand.Int())
			orgId, err := cli.CreateOrg(context.Background(), orgName)
			require.NoError(t, err)

			dashboardId, err := cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{"title": "Dashboard 1"}, true)
			require.NoError(t, err)
			require.NotEqual(t, DashboardID{}, dashboardId)

			updateDashboardId, err := cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{"title": "Dashboard 1 v2", "uid": dashboardId.String()}, true)
			require.NoError(t, err)
			require.Equal(t, dashboardId, updateDashboardId)
		})

		t.Run("ExistentDashboard/WithoutVersion/WithoutUID", func(t *testing.T) {
			orgName := fmt.Sprintf("foo-%v", rand.Int())
			orgId, err := cli.CreateOrg(context.Background(), orgName)
			require.NoError(t, err)

			dashboardId, err := cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{"title": "Dashboard 1"}, true)
			require.NoError(t, err)
			require.NotEqual(t, DashboardID{}, dashboardId)

			updateDashboardId, err := cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{"title": "Dashboard 1"}, true)
			require.NoError(t, err)
			require.Equal(t, dashboardId, updateDashboardId)
		})
	})
}

func TestIntegGetDashboardByUID(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cfg := config.GrafanaFromEnv()
	cli := ProvideClient(cfg)

	t.Run("NonExistentDashboard", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		dashboard, err := cli.GetDashboardByUID(context.Background(), orgId, DashboardID{})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrGrafanaDashboardNotFound)
		require.Equal(t, Dashboard{}, dashboard)
	})

	t.Run("ExistentDashboard", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		dashboardID, err := cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{"title": "Dashboard 1"}, false)
		require.NoError(t, err)

		dashboard, err := cli.GetDashboardByUID(context.Background(), orgId, dashboardID)
		require.NoError(t, err)

		require.IsType(t, float64(0), dashboard.Dashboard["id"])
		delete(dashboard.Dashboard, "id")

		require.IsType(t, float64(0), dashboard.Dashboard["version"])
		delete(dashboard.Dashboard, "version")

		_, err = ParseDashboardID(dashboard.Dashboard["uid"].(string))
		require.NoError(t, err)
		delete(dashboard.Dashboard, "uid")

		require.WithinDuration(t, time.Now(), dashboard.Metadata.Created, 2*time.Second)
		dashboard.Metadata.Created = time.Time{}
		require.WithinDuration(t, time.Now(), dashboard.Metadata.Updated, 2*time.Second)
		dashboard.Metadata.Updated = time.Time{}

		require.Equal(t, Dashboard{
			Dashboard: map[string]any{
				"title": "Dashboard 1",
			},
			Metadata: DashboardMetadata{
				AnnotationsPermissions: struct {
					dashboard struct {
						canAdd    bool
						canDelete bool
						canEdit   bool
					}
					organization struct {
						canAdd    bool
						canDelete bool
						canEdit   bool
					}
				}{},
				CanAdmin:               true,
				CanDelete:              true,
				CanEdit:                true,
				CanSave:                true,
				CanStar:                true,
				Created:                time.Time{},
				CreatedBy:              cfg.User.ExposeSecret(),
				Expires:                time.Time{},
				FolderId:               0,
				FolderTitle:            "General",
				FolderUid:              FolderID{},
				FolderUrl:              "",
				HasAcl:                 false,
				IsFolder:               false,
				IsSnapshot:             false,
				IsStarred:              false,
				Provisioned:            false,
				ProvisionedExternalId:  "",
				PublicDashboardEnabled: false,
				PublicDashboardUid:     "",
				Slug:                   "dashboard-1",
				Type:                   "db",
				Updated:                time.Time{},
				UpdatedBy:              cfg.User.ExposeSecret(),
				Url:                    "/d/" + dashboardID.String() + "/dashboard-1",
				Version:                1,
			},
		}, dashboard)
	})
}

func TestIntegDeleteDashboard(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cfg := config.GrafanaFromEnv()
	cli := ProvideClient(cfg)

	t.Run("NonExistentDashboard", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		err = cli.DeleteDashboardByUID(context.Background(), orgId, DashboardID(uuid.New()))
		require.Error(t, err)
		require.ErrorIs(t, err, ErrGrafanaDashboardNotFound)
	})

	t.Run("NonExistentDashboard", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		dashboardId, err := cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{"title": "Dashboard 1"}, true)
		require.NoError(t, err)
		require.NotEqual(t, DashboardID{}, dashboardId)

		err = cli.DeleteDashboardByUID(context.Background(), orgId, dashboardId)
		require.NoError(t, err)
	})
}

func TestIntegSearchDashboards(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	cfg := config.GrafanaFromEnv()
	cli := ProvideClient(cfg)

	t.Run("NoDashboard", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		results, err := cli.SearchDashboards(context.Background(), orgId, 100, 1)
		require.NoError(t, err)
		require.Len(t, results, 0)
	})

	t.Run("SingleDashboard", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		dashboardId, err := cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{"title": "Dashboard 1"}, true)
		require.NoError(t, err)

		results, err := cli.SearchDashboards(context.Background(), orgId, 100, 1)
		require.NoError(t, err)
		require.Len(t, results, 1)
		require.Equal(t, dashboardId, results[0].Uid)
		require.Equal(t, "Dashboard 1", results[0].Title)
	})

	t.Run("MultipleDashboard", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		var expectedSearchResults []SearchDashboardResult
		for i := 0; i < 10; i++ {
			dashboardTitle := fmt.Sprintf("Dashboard %v", i)
			dashboardId, err := cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{"title": dashboardTitle}, true)
			require.NoError(t, err)

			expectedSearchResults = append(expectedSearchResults, SearchDashboardResult{dashboardId, dashboardTitle})
		}

		results, err := cli.SearchDashboards(context.Background(), orgId, 100, 1)
		require.NoError(t, err)
		require.Len(t, results, len(expectedSearchResults))
		require.Equal(t, expectedSearchResults, results)
	})

	t.Run("MultiplePage", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		var expectedSearchResults []SearchDashboardResult
		for i := 0; i < 10; i++ {
			dashboardTitle := fmt.Sprintf("Dashboard %v", i)
			dashboardId, err := cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{"title": dashboardTitle}, true)
			require.NoError(t, err)

			expectedSearchResults = append(expectedSearchResults, SearchDashboardResult{dashboardId, dashboardTitle})
		}

		// Fetch first page.
		page1, err := cli.SearchDashboards(context.Background(), orgId, 5, 1)
		require.NoError(t, err)
		require.Len(t, page1, 5)

		// Fetch second page.
		page2, err := cli.SearchDashboards(context.Background(), orgId, 5, 2)
		require.NoError(t, err)
		require.Len(t, page2, 5)

		var searchResults []SearchDashboardResult
		searchResults = append(searchResults, page1...)
		searchResults = append(searchResults, page2...)

		require.Equal(t, expectedSearchResults, searchResults)
	})

	t.Run("NonExistentPage", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		_, err = cli.CreateUpdateDashboard(context.Background(), orgId, FolderID{}, map[string]any{"title": "Dashboard 1"}, true)
		require.NoError(t, err)

		results, err := cli.SearchDashboards(context.Background(), orgId, 100, 9)
		require.NoError(t, err)
		require.Len(t, results, 0)
	})
}
