package secret

import (
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
