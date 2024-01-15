package users

import (
	"context"
	"testing"
	"time"

	"github.com/prismelabs/prismeanalytics/internal/secret"
	"github.com/prismelabs/prismeanalytics/internal/testutils"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService(t *testing.T) {
	t.Run("CreateUser", func(t *testing.T) {
		ctx := context.Background()

		t.Run("UserAlreadyExists", func(t *testing.T) {
			// Setup store mock.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := NewMockStore(ctrl)

			// Setup service.
			service := newService(store)

			username := testutils.Must(NewUserName)("foo")
			email := testutils.Must(NewEmail)("foo@example.com")

			// Expect mock call.
			store.EXPECT().InsertUser(
				ctx,
				gomock.Any(), // user id
				username,
				email,
				gomock.Any(), // password
			).Return(ErrUserAlreadyExists)

			// Create user.
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
			// Setup store mock.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := NewMockStore(ctrl)

			// Setup service.
			service := newService(store)

			username := testutils.Must(NewUserName)("foo")
			email := testutils.Must(NewEmail)("foo@example.com")

			// Expect mock call.
			store.EXPECT().InsertUser(
				ctx,
				gomock.Any(), // user id
				username,
				email,
				gomock.Any(), // password
			).Return(nil)

			// Create user.
			userId, err := service.CreateUser(ctx, CreateCmd{
				UserName: username,
				Email:    email,
				Password: testutils.Must(NewPassword)(secret.New("p4ssW0rd!")),
			})
			require.NoError(t, err)
			require.NotEqual(t, UserId{}, userId)
		})
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		ctx := context.Background()

		t.Run("NonExistentUser", func(t *testing.T) {
			// Setup store mock.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := NewMockStore(ctrl)

			// Setup service.
			service := newService(store)

			email := testutils.Must(NewEmail)("foo@example.org")

			// Expect mock call.
			store.EXPECT().SelectUserByEmail(ctx, email).Return(User{}, ErrUserNotFound)

			// Retrieve user.
			user, err := service.GetUserByEmail(ctx, email)
			require.Error(t, err)
			require.ErrorIs(t, err, ErrUserNotFound)
			require.Equal(t, User{}, user)
		})

		t.Run("ExistentUser", func(t *testing.T) {
			// Setup store mock.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := NewMockStore(ctrl)

			// Setup service.
			service := newService(store)

			email := testutils.Must(NewEmail)("foo@example.org")

			// Expect mock call.
			expectedUser := User{
				Id:        NewUserId(),
				Email:     email,
				Password:  testutils.Must(NewPassword)(secret.New("s3cretPassw0rd")),
				Name:      testutils.Must(NewUserName)("foo"),
				CreatedAt: time.Now(),
			}
			store.EXPECT().SelectUserByEmail(ctx, email).Return(expectedUser, nil)

			// Retrieve user.
			user, err := service.GetUserByEmail(ctx, email)
			require.NoError(t, err)
			require.Equal(t, expectedUser, user)
		})
	})
}
