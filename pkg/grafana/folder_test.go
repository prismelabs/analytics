package grafana

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/prismelabs/analytics/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func TestIntegCreateFolder(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	cli := ProvideClient(cfg)

	t.Run("NonExistentTitle", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		folder, err := cli.CreateFolder(context.Background(), orgId, "Folder 1")
		require.NoError(t, err)
		require.Equal(t, folder.Title, "Folder 1", "%+v")
		require.NotEqual(t, int64(0), folder.Id)
		require.NotEqual(t, uuid.UUID{}, folder.Uid)
	})

	t.Run("ExistentTitle", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		_, err = cli.CreateFolder(context.Background(), orgId, "Folder 1")
		require.NoError(t, err)

		// Create file again.
		_, err = cli.CreateFolder(context.Background(), orgId, "Folder 1")
		require.NoError(t, err)
	})
}

func TestIntegListFolders(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	cli := ProvideClient(cfg)

	t.Run("NoFolder", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		folders, err := cli.ListFolders(context.Background(), orgId, 100, 1)
		require.NoError(t, err)
		require.Len(t, folders, 0)
	})

	t.Run("SingleFolder", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		folder, err := cli.CreateFolder(context.Background(), orgId, "Folder 1")
		require.NoError(t, err)

		folders, err := cli.ListFolders(context.Background(), orgId, 100, 1)
		require.NoError(t, err)
		require.Len(t, folders, 1)
		require.Equal(t, folder, folders[0])
	})

	t.Run("MultipleFolder", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		var expectedFolders []Folder
		for i := 0; i < 10; i++ {
			folder, err := cli.CreateFolder(context.Background(), orgId, fmt.Sprintf("Folder %v", i))
			require.NoError(t, err)

			expectedFolders = append(expectedFolders, folder)
		}

		folders, err := cli.ListFolders(context.Background(), orgId, 100, 1)
		require.NoError(t, err)
		require.Len(t, folders, len(expectedFolders))
		for _, expected := range expectedFolders {
			found := false
			for _, actual := range folders {
				if expected.Uid.String() == actual.Uid.String() {
					found = true
					break
				}
			}

			require.Truef(t, found, "folder not found: %+v", expected)
		}
	})

	t.Run("MultiplePage", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		var expectedFolders []Folder
		for i := 0; i < 10; i++ {
			folder, err := cli.CreateFolder(context.Background(), orgId, fmt.Sprintf("Folder %v", i))
			require.NoError(t, err)

			expectedFolders = append(expectedFolders, folder)
		}

		// Fetch first page.
		page1, err := cli.ListFolders(context.Background(), orgId, 5, 1)
		require.NoError(t, err)
		require.Len(t, page1, 5)

		// Fetch second page.
		page2, err := cli.ListFolders(context.Background(), orgId, 5, 2)
		require.NoError(t, err)
		require.Len(t, page2, 5)

		var folders []Folder
		folders = append(folders, page1...)
		folders = append(folders, page2...)

		for _, expected := range expectedFolders {
			found := false
			for _, actual := range folders {
				if expected.Uid.String() == actual.Uid.String() {
					found = true
					break
				}
			}

			require.Truef(t, found, "folder not found: %+v", expected)
		}
	})

	t.Run("NonExistentPage", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		folders, err := cli.ListFolders(context.Background(), orgId, 100, 4)
		require.NoError(t, err)
		require.Len(t, folders, 0)
	})
}

func TestIntegListFolderPermissins(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	cli := ProvideClient(cfg)

	t.Run("NonExistentFolder", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		perms, err := cli.GetFolderPermissions(context.Background(), orgId, Uid{})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrGrafanaFolderNotFound)
		require.Len(t, perms, 0)
	})

	t.Run("DefaultPermissions", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		folder, err := cli.CreateFolder(context.Background(), orgId, "Folder 1")
		require.NoError(t, err)

		expectedPerms := []FolderPermission{
			{
				Permission: FolderPermissionLevelAdmin,
				UserId:     1,
			},
			{
				Permission: FolderPermissionLevelEdit,
				Role:       RoleEditor,
			},
			{
				Permission: FolderPermissionLevelView,
				Role:       RoleViewer,
			},
		}

		perms, err := cli.GetFolderPermissions(context.Background(), orgId, folder.Uid)
		require.NoError(t, err)
		require.Len(t, perms, len(expectedPerms))

		for _, expected := range expectedPerms {
			found := false
			for _, actual := range perms {
				if expected == actual {
					found = true
					break
				}
			}

			require.Truef(t, found, "folder permission %+v not found", expected)
		}
	})

	t.Run("CustomPermissions", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		folder, err := cli.CreateFolder(context.Background(), orgId, "Folder 1")
		require.NoError(t, err)

		expectedPerms := []FolderPermission{
			{
				Permission: FolderPermissionLevelEdit,
				Role:       RoleViewer,
			},
			{
				Permission: FolderPermissionLevelView,
				Role:       RoleEditor,
			},
		}

		err = cli.SetFolderPermissions(context.Background(), orgId, folder.Uid, expectedPerms...)
		require.NoError(t, err)

		perms, err := cli.GetFolderPermissions(context.Background(), orgId, folder.Uid)
		require.NoError(t, err)
		require.Len(t, perms, len(expectedPerms))

		for _, expected := range expectedPerms {
			found := false
			for _, actual := range perms {
				if expected == actual {
					found = true
					break
				}
			}

			require.Truef(t, found, "folder permission %+v not found", expected)
		}
	})
}

func TestIntegSetFolderPermissions(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	cli := ProvideClient(cfg)

	type testCase struct {
		name             string
		folderPerms      []FolderPermission
		expectedErrorMsg string
	}

	testCases := []testCase{
		{
			name: "NonExistentUserId",
			folderPerms: []FolderPermission{
				{
					Permission: FolderPermissionLevelAdmin,
					UserId:     999999,
				},
			},
			expectedErrorMsg: "Failed to create permission",
		},
		{
			name: "NonExistentTeamId",
			folderPerms: []FolderPermission{
				{
					Permission: FolderPermissionLevelAdmin,
					TeamId:     999999,
				},
			},
			expectedErrorMsg: "Failed to create permission",
		},
		{
			name: "UserPermAndRolePerm",
			folderPerms: []FolderPermission{
				{
					Permission: FolderPermissionLevelAdmin,
					UserId:     1,
				},
				{
					Permission: FolderPermissionLevelEdit,
					Role:       RoleEditor,
				},
			},
		},
		{
			name:        "Empty",
			folderPerms: []FolderPermission{},
		},
	}

	for _, tcase := range testCases {
		t.Run(tcase.name, func(t *testing.T) {
			orgName := fmt.Sprintf("foo-%v", rand.Int())
			orgId, err := cli.CreateOrg(context.Background(), orgName)
			require.NoError(t, err)

			folder, err := cli.CreateFolder(context.Background(), orgId, "Folder 1")
			require.NoError(t, err)

			err = cli.SetFolderPermissions(context.Background(), orgId, folder.Uid, tcase.folderPerms...)
			if tcase.expectedErrorMsg != "" {
				require.Error(t, err)
				require.Regexp(t, tcase.expectedErrorMsg, err)
				return
			}
			require.NoError(t, err)

			perms, err := cli.GetFolderPermissions(context.Background(), orgId, folder.Uid)
			require.NoError(t, err)
			require.Len(t, perms, len(tcase.folderPerms))

			for _, expected := range tcase.folderPerms {
				found := false
				for _, actual := range perms {
					if expected == actual {
						found = true
						break
					}
				}

				require.Truef(t, found, "folder permission %+v not found", expected)
			}
		})
	}

	t.Run("DuplicateRole", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		folder, err := cli.CreateFolder(context.Background(), orgId, "Folder 1")
		require.NoError(t, err)

		expectedPerms := []FolderPermission{
			{
				Permission: FolderPermissionLevelEdit,
				Role:       RoleViewer,
			},
			{
				Permission: FolderPermissionLevelView,
				Role:       RoleViewer,
			},
		}

		err = cli.SetFolderPermissions(context.Background(), orgId, folder.Uid, expectedPerms...)
		require.NoError(t, err)

		perms, err := cli.GetFolderPermissions(context.Background(), orgId, folder.Uid)
		require.NoError(t, err)
		require.Equal(t, []FolderPermission{{
			Permission: FolderPermissionLevelView,
			Role:       RoleViewer,
		}}, perms)
	})

	t.Run("NoPermission/AdminCanStillEditPermissions", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		folder, err := cli.CreateFolder(context.Background(), orgId, "Folder 1")
		require.NoError(t, err)

		err = cli.SetFolderPermissions(context.Background(), orgId, folder.Uid)
		require.NoError(t, err)

		// Change permissions again even though we didn't allow admin user to
		// edit permission in above call.
		err = cli.SetFolderPermissions(context.Background(), orgId, folder.Uid, FolderPermission{
			Permission: FolderPermissionLevelAdmin,
			Role:       RoleEditor,
			TeamId:     0,
			UserId:     0,
		})
		require.NoError(t, err)
	})
}

func TestIntegDeleteFolder(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	cli := ProvideClient(cfg)

	t.Run("NonExistentFolder", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		err = cli.DeleteFolder(context.Background(), orgId, Uid{})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrGrafanaFolderNotFound)
	})

	t.Run("ExistentFolder", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		folder, err := cli.CreateFolder(context.Background(), orgId, "Folder 1")
		require.NoError(t, err)

		err = cli.DeleteFolder(context.Background(), orgId, folder.Uid)
		require.NoError(t, err)
	})
}

func TestIntegSearchFolders(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	var cfg Config
	testutils.ConfigueLoad(t, &cfg)
	cli := ProvideClient(cfg)

	t.Run("NoFolder", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		results, err := cli.SearchFolders(context.Background(), orgId, 100, 1, "")
		require.NoError(t, err)
		require.Len(t, results, 0)
	})

	t.Run("SingleFolder", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		folder, err := cli.CreateFolder(context.Background(), orgId, "Foo")
		require.NoError(t, err)

		results, err := cli.SearchFolders(context.Background(), orgId, 100, 1, "")
		require.NoError(t, err)
		require.Len(t, results, 1)
		require.Equal(t, folder.Uid, results[0].Uid)
		require.Equal(t, "Foo", results[0].Title)
	})

	t.Run("MultipleFolder", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		var expectedSearchResults []SearchFolderResult
		for i := 0; i < 10; i++ {
			folderName := fmt.Sprintf("Folder %v", i)
			folder, err := cli.CreateFolder(context.Background(), orgId, folderName)
			require.NoError(t, err)

			expectedSearchResults = append(expectedSearchResults, SearchFolderResult{
				Uid:   folder.Uid,
				Title: folderName,
				Url:   "", // Checked separately.
			})
		}

		searchResults, err := cli.SearchFolders(context.Background(), orgId, 100, 1, "")
		require.NoError(t, err)
		require.Len(t, searchResults, len(expectedSearchResults))

		// result.Url is random so we check it here.
		for i, result := range searchResults {
			require.Contains(t, result.Url, "/f/")
			require.Contains(t, result.Url, strings.ReplaceAll(strings.ToLower(result.Title), " ", "-"))
			searchResults[i].Url = ""
		}
		require.Equal(t, expectedSearchResults, searchResults)
	})

	t.Run("MultiplePage", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		var expectedSearchResults []SearchFolderResult
		for i := 0; i < 10; i++ {
			folderName := fmt.Sprintf("Folder %v", i)
			folder, err := cli.CreateFolder(context.Background(), orgId, folderName)
			require.NoError(t, err)

			expectedSearchResults = append(expectedSearchResults, SearchFolderResult{
				Uid:   folder.Uid,
				Title: folderName,
				Url:   "",
			})
		}

		// Fetch first page.
		page1, err := cli.SearchFolders(context.Background(), orgId, 5, 1, "")
		require.NoError(t, err)
		require.Len(t, page1, 5)

		// Fetch second page.
		page2, err := cli.SearchFolders(context.Background(), orgId, 5, 2, "")
		require.NoError(t, err)
		require.Len(t, page2, 5)

		var searchResults []SearchFolderResult
		searchResults = append(searchResults, page1...)
		searchResults = append(searchResults, page2...)

		// result.Url is random so we check it here.
		for i, result := range searchResults {
			require.Contains(t, result.Url, "/f/")
			require.Contains(t, result.Url, strings.ReplaceAll(strings.ToLower(result.Title), " ", "-"))
			searchResults[i].Url = ""
		}
		require.Equal(t, expectedSearchResults, searchResults)
	})

	t.Run("NoResult", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		_, err = cli.CreateFolder(context.Background(), orgId, "Folder 1")
		require.NoError(t, err)

		results, err := cli.SearchFolders(context.Background(), orgId, 100, 9, "Non existent dashboard")
		require.NoError(t, err)
		require.Len(t, results, 0)
	})

	t.Run("NonExistentPage", func(t *testing.T) {
		orgName := fmt.Sprintf("foo-%v", rand.Int())
		orgId, err := cli.CreateOrg(context.Background(), orgName)
		require.NoError(t, err)

		_, err = cli.CreateFolder(context.Background(), orgId, "Folder 1")
		require.NoError(t, err)

		results, err := cli.SearchFolders(context.Background(), orgId, 100, 9, "")
		require.NoError(t, err)
		require.Len(t, results, 0)
	})
}
