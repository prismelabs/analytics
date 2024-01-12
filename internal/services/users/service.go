package users

import (
	"context"
	"fmt"

	"github.com/prismelabs/prismeanalytics/internal/models"
	"github.com/prismelabs/prismeanalytics/internal/secret"
	"golang.org/x/crypto/bcrypt"
)

// Service define user management service.
type Service interface {
	CreateUser(context.Context, CreateCmd) (models.UserId, error)
}

// ProvideService define a wire provider for user Service.
func ProvideService(store Store) Service {
	return service{store}
}

type service struct {
	store Store
}

// CreateUser implements Service.
func (s service) CreateUser(ctx context.Context, cmd CreateCmd) (models.UserId, error) {
	uid := models.NewUserId()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cmd.Password.ExposeSecret()), bcrypt.DefaultCost)
	if err != nil {
		return models.UserId{}, fmt.Errorf("failed to hash password: %w", err)
	}

	err = s.store.InsertUser(ctx, uid, cmd.UserName, cmd.Email, PasswordHash(secret.New(hashedPassword)))
	if err != nil {
		return models.UserId{}, fmt.Errorf("failed to insert user in store: %w", err)
	}

	// TODO: sent verification email and implement email verification.

	return uid, nil
}
