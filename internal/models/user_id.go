package models

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrUserIdIsNotAUuidV4 = errors.New("user id is not a uuid version 4")
)

// UserId define a unique user identifier.
// Every UUIDv4 is a valid user id.
type UserId struct {
	value uuid.UUID
}

// NewUserId generates a new random user id.
func NewUserId() UserId {
	return UserId{uuid.New()}
}

// ParseUserId parses the given value as a user id.
// An error is returned if the id doesn't satisify user id requirements.
func ParseUserId(value string) (UserId, error) {
	uid, err := uuid.Parse(value)
	if err != nil {
		return UserId{}, fmt.Errorf("%w: %w", ErrUserIdIsNotAUuidV4, err)
	}
	if uid.Version() != 4 {
		return UserId{}, ErrUserIdIsNotAUuidV4
	}

	return UserId{uid}, nil
}

// String implements fmt.Stringer.
func (uid UserId) String() string {
	return uid.value.String()
}

// Scan implements sql.Scanner.
func (uid *UserId) Scan(src any) error {
	nid := uuid.NullUUID{}
	err := nid.Scan(src)
	if err != nil {
		return err
	}
	if !nid.Valid {
		return fmt.Errorf("user ID can't be NULL")
	}
	if nid.UUID.Version() != 4 {
		return fmt.Errorf("invalid user id, UUID v4 expected, got %d", nid.UUID.Version())
	}

	copy(uid.value[:], nid.UUID[:])

	return nil
}

// Value implements driver.Valuer.
func (uid UserId) Value() (driver.Value, error) {
	return uid.String(), nil
}
