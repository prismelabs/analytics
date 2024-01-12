package users

import (
	"context"
	"database/sql"

	"github.com/prismelabs/prismeanalytics/internal/models"
	"github.com/prismelabs/prismeanalytics/internal/postgres"
	"github.com/prismelabs/prismeanalytics/internal/secret"
)

// Store define a user store.
type Store interface {
	InsertUser(context.Context, models.UserId, models.UserName, models.Email, PasswordHash) error
}

// ProvideStore define a wire provider for user store.
func ProvideStore(pg postgres.Pg) Store {
	return store{pg.DB}
}

type store struct {
	db *sql.DB
}

// InsertUser implements Store.
func (s store) InsertUser(ctx context.Context, userId models.UserId, userName models.UserName, email models.Email, passwordHash PasswordHash) error {
	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO users VALUES ($1, $2, $3, $4, NOW())",
		userId,
		userName,
		email,
		secret.Secret[[]byte](passwordHash).ExposeSecret(),
	)
	if err != nil {
		return err
	}

	return nil
}
