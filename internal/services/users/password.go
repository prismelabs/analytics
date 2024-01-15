package users

import (
	"errors"
	"fmt"

	"github.com/prismelabs/prismeanalytics/internal/secret"
)

var (
	ErrPasswordTooShort = errors.New("password too short")
	ErrPasswordTooLong  = errors.New("password too long")
)

// Password define a secret with a length between 6 and 72 bytes.
type Password struct {
	value secret.Secret[string]
}

// NewPassword returns a new password from the given string.
// An error is returned if the given value doesn't satisfy password requirements.
func NewPassword(value secret.Secret[string]) (Password, error) {
	secretLength := len(value.ExposeSecret())
	if secretLength < 6 {
		return Password{}, ErrPasswordTooShort
	} else if secretLength > 72 {
		return Password{}, ErrPasswordTooLong
	}

	return Password{value}, nil
}

// ExposeSecret returns the underlying secret.
// It is best practice to never store the secret and expose it only when needed.
func (p Password) ExposeSecret() string {
	return p.value.ExposeSecret()
}

// Scan implements sql.Scanner.
func (p *Password) Scan(src any) error {
	if t, ok := src.(string); ok {
		p.value = secret.New(t)
		return nil
	}
	return fmt.Errorf("cannot scan %T into Domain", src)
}
