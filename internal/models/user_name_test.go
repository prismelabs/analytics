package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewUserName(t *testing.T) {
	t.Run("TooShort", func(t *testing.T) {
		un, err := NewUserName("🏳️‍🌈")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUserNameTooShort)
		require.Equal(t, UserName{}, un)
	})

	t.Run("TooLong", func(t *testing.T) {
		un, err := NewUserName(strings.Repeat("Foo ", 16)) // length of 64
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUserNameTooLong)
		require.Equal(t, UserName{}, un)
	})

	t.Run("Valid", func(t *testing.T) {
		un, err := NewUserName("Foo")
		require.NoError(t, err)
		require.Equal(t, "Foo", un.String())
	})
}
