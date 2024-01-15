package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/prismelabs/prismeanalytics/internal/secret"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

// store define a user store.
//
//go:generate mockgen -source store.go -destination store_mock_test.go -package users -mock_names store=MockStore store
type store interface {
	InsertUser(context.Context, UserId, UserName, Email, secret.Secret[[]byte]) error
	SelectUserByEmail(context.Context, Email) (User, error)
}

type pgStore struct {
	db *sql.DB
}

// InsertUser implements Store.
func (pgs pgStore) InsertUser(ctx context.Context, userId UserId, userName UserName, email Email, passwordHash secret.Secret[[]byte]) error {
	_, err := pgs.db.ExecContext(
		ctx,
		"INSERT INTO users VALUES ($1, $2, $3, $4, NOW())",
		userId,
		userName,
		email,
		secret.Secret[[]byte](passwordHash).ExposeSecret(),
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
			return ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

// SelectUserByEmail implements store.
func (pgs pgStore) SelectUserByEmail(ctx context.Context, email Email) (User, error) {
	row := pgs.db.QueryRowContext(
		ctx,
		"SELECT id, name, password, created_at FROM users WHERE email = $1",
		email,
	)

	result := User{
		Email: email,
	}

	err := row.Scan(&result.Id, &result.Name, &result.Password, &result.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("failed to scan user: %w", err)
	}

	return result, nil
}
