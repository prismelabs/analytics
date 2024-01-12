package users

import (
	"github.com/prismelabs/prismeanalytics/internal/models"
	"github.com/prismelabs/prismeanalytics/internal/secret"
)

type CreateCmd struct {
	UserName models.UserName
	Email    models.Email
	Password models.Password
}

// PasswordHash define a hashed password.
type PasswordHash secret.Secret[[]byte]
