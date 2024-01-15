package users

import (
	"context"
	"fmt"

	"github.com/prismelabs/prismeanalytics/internal/postgres"
	"github.com/prismelabs/prismeanalytics/internal/secret"
	"golang.org/x/crypto/bcrypt"
)

// Service define user management service.
type Service interface {
	CreateUser(context.Context, CreateCmd) (UserId, error)
	GetUserByEmail(context.Context, Email) (User, error)
}

// ProvideService define a wire provider for user Service.
func ProvideService(pg postgres.Pg) Service {
	return newService(pgStore{pg.DB})
}

func newService(store store) service {
	return service{store}
}

type service struct {
	store store
}

// CreateUser implements Service.
func (s service) CreateUser(ctx context.Context, cmd CreateCmd) (UserId, error) {
	uid := NewUserId()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cmd.Password.ExposeSecret()), bcrypt.DefaultCost)
	if err != nil {
		return UserId{}, fmt.Errorf("failed to hash password: %w", err)
	}

	err = s.store.InsertUser(ctx, uid, cmd.UserName, cmd.Email, secret.New(hashedPassword))
	if err != nil {
		return UserId{}, fmt.Errorf("failed to insert user in store: %w", err)
	}

	// TODO: sent verification email and implement email verification.

	return uid, nil
}

// GetUserByEmail implements Service.
func (s service) GetUserByEmail(ctx context.Context, email Email) (User, error) {
	user, err := s.store.SelectUserByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("failed to select user by email: %w", err)
	}

	return user, nil
}
