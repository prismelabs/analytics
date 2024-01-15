package auth

import (
	"context"
	"testing"
	"time"

	"github.com/prismelabs/prismeanalytics/internal/secret"
	"github.com/prismelabs/prismeanalytics/internal/services/users"
	"github.com/prismelabs/prismeanalytics/internal/testutils"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService(t *testing.T) {
	t.Run("AuthenticateByPassword", func(t *testing.T) {
		ctx := context.Background()

		t.Run("UserNotFound", func(t *testing.T) {
			// Setup user service mock.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userService := NewMockUserService(ctrl)

			// Setup auth service.
			authService := ProvideService(userService)

			email := testutils.Must(users.NewEmail)("foo@example.com")
			password := secret.New("s3cretP4ssw0rd")

			// Expect user service mock call.
			userService.EXPECT().GetUserByEmail(ctx, email).
				Return(users.User{}, users.ErrUserNotFound)

			// Authenticate user.
			user, err := authService.AuthenticateByPassword(ctx, email, password)
			require.Error(t, err)
			require.ErrorIs(t, err, ErrInvalidCredentials)
			require.Equal(t, users.User{}, user)
		})

		t.Run("UnexpectedError", func(t *testing.T) {
			// Setup user service mock.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			userService := NewMockUserService(ctrl)

			// Setup auth service.
			authService := ProvideService(userService)

			email := testutils.Must(users.NewEmail)("foo@example.com")
			password := secret.New("s3cretP4ssw0rd")
			expectedUser := users.User{
				Id:        users.NewUserId(),
				Email:     email,
				Password:  secret.New("$2a$10$2mpwHvtlaF5JWTgpG7edwu1KfB0LP5vPoA9B.2ANPbGOIuQgqx3nW"),
				Name:      testutils.Must(users.NewUserName)("foo"),
				CreatedAt: time.Now(),
			}

			// Expect user service mock call.
			userService.EXPECT().GetUserByEmail(ctx, email).
				Return(expectedUser, nil)

			// Authenticate user.
			user, err := authService.AuthenticateByPassword(ctx, email, password)
			require.NoError(t, err)
			require.Equal(t, expectedUser, user)
		})
	})
}
