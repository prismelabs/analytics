package models

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEmail(t *testing.T) {
	t.Run("InvalidEmail", func(t *testing.T) {
		email, err := NewEmail("")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrEmailInvalid)
		require.Equal(t, Email{}, email)
	})

	t.Run("ValidEmail", func(t *testing.T) {
		email, err := NewEmail("foo@example.com")
		require.NoError(t, err)
		require.Equal(t, "foo@example.com", email.String())
	})
}
