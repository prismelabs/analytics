package orgs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prismelabs/prismeanalytics/internal/services/users"
	"github.com/prismelabs/prismeanalytics/internal/testutils"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateOrg", func(t *testing.T) {
		t.Run("NameAlreadyTaken", func(t *testing.T) {
			// Setup store mock.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := NewMockStore(ctrl)

			// Setup service.
			orgService := newService(store)

			// Random user id.
			userId := users.NewUserId()

			orgName := testutils.Must(NewOrgName)("foo's org")

			// Expect mock calls.
			store.EXPECT().InsertOrg(ctx, OrganizationMatcher{
				Id:           gomock.Any(),
				OwnerId:      gomock.AnyOf(userId),
				GrafanaOrgId: gomock.AnyOf(GrafanaOrgId(0)),
				Name:         gomock.AnyOf(orgName),
				CreatedAt:    gomock.Any(),
			}).Return(ErrOrgNameAlreadyTaken)

			// Create org.
			org, err := orgService.CreateOrg(ctx, userId, orgName)
			require.Error(t, err)
			require.ErrorIs(t, err, ErrOrgNameAlreadyTaken)
			require.Equal(t, Organization{}, org)
		})

		t.Run("Valid", func(t *testing.T) {
			// Setup store mock.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := NewMockStore(ctrl)

			// Setup service.
			orgService := newService(store)

			// Random user id.
			userId := users.NewUserId()

			orgName := testutils.Must(NewOrgName)("foo's org")

			// Expect mock calls.
			store.EXPECT().InsertOrg(ctx, OrganizationMatcher{
				Id:           gomock.Any(),
				OwnerId:      gomock.AnyOf(userId),
				GrafanaOrgId: gomock.AnyOf(GrafanaOrgId(0)),
				Name:         gomock.AnyOf(orgName),
				CreatedAt:    gomock.Any(),
			}).Return(nil)

			// Create org.
			org, err := orgService.CreateOrg(ctx, userId, orgName)
			require.NoError(t, err)
			require.NotEqual(t, OrgId{}, org.Id)
			require.Equal(t, userId, org.OwnerId)
			require.Equal(t, GrafanaOrgId(0), org.GrafanaOrgId)
			require.Equal(t, org.Name, orgName)
			require.WithinDuration(t, time.Now(), org.CreatedAt, 3*time.Second)
		})
	})

	t.Run("GetOrgById", func(t *testing.T) {
		t.Run("StoreError", func(t *testing.T) {
			storeErr := errors.New("unexpected error")

			// Setup store mock.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := NewMockStore(ctrl)

			// Setup service.
			orgService := newService(store)

			// Random org id.
			orgId := NewOrgId()

			// Expect mock calls.
			store.EXPECT().SelectOrgById(ctx, orgId).Return(Organization{}, storeErr)

			// Create org.
			org, err := orgService.GetOrgById(ctx, orgId)
			require.Error(t, err)
			require.ErrorIs(t, err, storeErr)
			require.Equal(t, Organization{}, org)
		})

		t.Run("OrgNotFound", func(t *testing.T) {
			// Setup store mock.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := NewMockStore(ctrl)

			// Setup service.
			orgService := newService(store)

			// Random org id.
			orgId := NewOrgId()

			// Expect mock calls.
			store.EXPECT().SelectOrgById(ctx, orgId).Return(Organization{}, ErrOrgNotFound)

			// Create org.
			org, err := orgService.GetOrgById(ctx, orgId)
			require.Error(t, err)
			require.ErrorIs(t, err, ErrOrgNotFound)
			require.Equal(t, Organization{}, org)
		})

		t.Run("OrgFound", func(t *testing.T) {
			// Setup store mock.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := NewMockStore(ctrl)

			// Setup service.
			orgService := newService(store)

			// Random org id.
			orgId := NewOrgId()

			expectedOrg := Organization{
				Id:           orgId,
				OwnerId:      users.NewUserId(),
				GrafanaOrgId: 0,
				Name:         testutils.Must(NewOrgName)("foo's org"),
				CreatedAt:    time.Now(),
			}

			// Expect mock calls.
			store.EXPECT().SelectOrgById(ctx, orgId).
				Return(expectedOrg, nil)

			// Create org.
			org, err := orgService.GetOrgById(ctx, orgId)
			require.NoError(t, err)
			require.Equal(t, expectedOrg, org)
		})
	})
}
