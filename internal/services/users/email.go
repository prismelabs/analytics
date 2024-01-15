package users

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"net/mail"
)

var (
	ErrEmailInvalid = errors.New("email invalid")
)

// Email define a valid email (RFC 5322) address of the form "foo@example.com".
type Email struct {
	value string
}

// NewEmail returns a new email from the given string.
// An error is returned if the given value isn't valid.
func NewEmail(value string) (Email, error) {
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return Email{}, fmt.Errorf("%w: %w", ErrEmailInvalid, err)
	}
	if addr.Name != "" {
		return Email{}, ErrEmailInvalid
	}
	if addr.Address != value {
		return Email{}, ErrEmailInvalid
	}

	return Email{value}, nil
}

// String implements fmt.Stringer.
func (e Email) String() string {
	return e.value
}

// Scan implements sql.Scanner.
func (e *Email) Scan(src any) error {
	if t, ok := src.(string); ok {
		e.value = t
		return nil
	}
	return fmt.Errorf("cannot scan %T into Domain", src)
}

// Value implements driver.Valuer.
func (e Email) Value() (driver.Value, error) {
	return e.String(), nil
}
