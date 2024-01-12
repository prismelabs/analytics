package models

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/rivo/uniseg"
)

var (
	ErrUserNameTooShort = errors.New("user name too short")
	ErrUserNameTooLong  = errors.New("user name too long")
)

// UserName define a user name.
// A user name is considered valid if it contains at least 3 non whitespace
// character.
type UserName struct {
	value string
}

// NewUserName returns a new user name from the given string.
// An error is returned if the given value doesn't satisfy user name requirements.
func NewUserName(value string) (UserName, error) {
	gcCount := uniseg.GraphemeClusterCount(value)
	if gcCount < 3 {
		return UserName{}, ErrUserNameTooShort
	} else if gcCount >= 64 {
		return UserName{}, ErrUserNameTooLong
	}

	return UserName{value}, nil
}

// String implements fmt.Stringer.
func (un UserName) String() string {
	return un.value
}

// Scan implements sql.Scanner.
func (un *UserName) Scan(src any) error {
	if t, ok := src.(string); ok {
		un.value = t
		return nil
	}
	return fmt.Errorf("cannot scan %T into Domain", src)
}

// Value implements driver.Valuer.
func (un UserName) Value() (driver.Value, error) {
	return un.String(), nil
}
