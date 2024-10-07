// Package secret contains utils related to secrets.
package secret

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// New wrapps the given secret in Secret type and returns it.
func New[T any](secret T) Secret[T] {
	return Secret[T]{secret}
}

// Secret define a wrapper of T to prevent unwanted exposure of a secret
// (in logs for example).
type Secret[T any] struct {
	value T
}

// ExposeSecret returns the underlying secret.
// It is best practice to never store the secret and expose it only when needed.
func (s Secret[T]) ExposeSecret() T {
	return s.value
}

// String implements fmt.Stringer.
func (s Secret[T]) String() string {
	typeName := reflect.TypeOf(s.value).String()
	return fmt.Sprintf("Secret[%v](******)", typeName)
}

// Scan implements sql.Scanner.
func (s *Secret[T]) Scan(src any) error {
	if t, ok := src.(T); ok {
		*s = New(t)
		return nil
	}
	return fmt.Errorf("cannot scan %T into Secret[%T]", src, *new(T))
}

// MarshalJSON implements json.Marshaler.
func (s Secret[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
