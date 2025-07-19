package options

import (
	"github.com/negrel/configue"
	"github.com/negrel/secrecy"
)

var _ configue.Value = Secret{}

// Secret defines a secret string option that implements configue.Option.
type Secret struct {
	*secrecy.SecretString
}

// Set implements configue.Value.
func (s Secret) Set(secret string) error {
	*s.SecretString = secrecy.NewSecretString(secrecy.UnsafeStringToBytes(secret))
	return nil
}

// String implements configue.Value.
func (s Secret) String() string {
	return ""
}
