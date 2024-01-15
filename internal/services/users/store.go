package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/prismelabs/prismeanalytics/internal/secret"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

// store define a user store.
//
//go:generate mockgen -source store.go -destination store_mock_test.go -package users -mock_names store=MockStore store
type store interface {
	InsertUser(context.Context, UserId, UserName, Email, secret.Secret[[]byte]) error
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
