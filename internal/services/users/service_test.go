package users

import (
	"context"
	"testing"

	"github.com/prismelabs/prismeanalytics/internal/secret"
	"github.com/prismelabs/prismeanalytics/internal/testutils"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService(t *testing.T) {
	t.Run("CreateUser", func(t *testing.T) {
		ctx := context.Background()

		t.Run("UserAlreadyExists", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := NewMockStore(ctrl)

			service := newService(store)

			username := testutils.Must(NewUserName)("foo")
			email := testutils.Must(NewEmail)("foo@example.com")

			store.EXPECT().InsertUser(
				ctx,
				gomock.Any(), // user id
				username,
				email,
				gomock.Any(), // password
			).Return(ErrUserAlreadyExists)

			userId, err := service.CreateUser(ctx, CreateCmd{
				UserName: username,
				Email:    email,
				Password: testutils.Must(NewPassword)(secret.New("p4ssW0rd!")),
			})
			require.Error(t, err)
			require.ErrorIs(t, err, ErrUserAlreadyExists)
			require.Equal(t, UserId{}, userId)
		})

		t.Run("EmailNotUsed", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := NewMockStore(ctrl)

			service := newService(store)

			username := testutils.Must(NewUserName)("foo")
			email := testutils.Must(NewEmail)("foo@example.com")

			store.EXPECT().InsertUser(
				ctx,
				gomock.Any(), // user id
				username,
				email,
				gomock.Any(), // password
			).Return(nil)

			userId, err := service.CreateUser(ctx, CreateCmd{
				UserName: username,
				Email:    email,
				Password: testutils.Must(NewPassword)(secret.New("p4ssW0rd!")),
			})
			require.NoError(t, err)
			require.NotEqual(t, UserId{}, userId)
		})
	})
}
