package saltmanager

import (
	"crypto/rand"
	"fmt"
)

type Salt [16]byte

// randomSalt generates a new crypto random salt.
func randomSalt() (Salt, error) {
	var salt Salt
	n, err := rand.Read(salt[:])
	if n != cap(salt) {
		return Salt{}, fmt.Errorf("failed to read %v random bytes: %v bytes read", cap(salt), n)
	}
	if err != nil {
		return Salt{}, err
	}

	return salt, nil
}

// Bytes convert salt to bytes.
func (s Salt) Bytes() []byte {
	return s[:]
}
