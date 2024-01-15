package auth

import (
	"context"
	"errors"

	"github.com/prismelabs/prismeanalytics/internal/secret"
	"github.com/prismelabs/prismeanalytics/internal/services/users"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("login or password incorrect")
)

// Service define authentication service.
type Service interface {
	AuthenticateByPassword(context.Context, users.Email, secret.Secret[string]) (users.User, error)
}

// ProvideService is a wire provider for authentication service.
func ProvideService(userService users.Service) Service {
	return service{userService}
}

type service struct {
	userService users.Service
}

// Authenticate implements Service.
func (s service) AuthenticateByPassword(ctx context.Context, email users.Email, password secret.Secret[string]) (users.User, error) {
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return users.User{}, ErrInvalidCredentials
		}

		return users.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password.ExposeSecret()), []byte(password.ExposeSecret()))
	if err != nil {
		return users.User{}, ErrInvalidCredentials
	}

	return user, nil
}
