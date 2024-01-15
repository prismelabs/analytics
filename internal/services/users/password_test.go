package users

import (
	"strings"
	"testing"

	"github.com/prismelabs/prismeanalytics/internal/secret"
	"github.com/stretchr/testify/require"
)

func TestPassword(t *testing.T) {
	t.Run("TooShort", func(t *testing.T) {
		passwd, err := NewPassword(secret.New("abcde"))
		require.Error(t, err)
		require.ErrorIs(t, err, ErrPasswordTooShort)
		require.Equal(t, Password{}, passwd)
	})

	t.Run("TooLong", func(t *testing.T) {
		passwd, err := NewPassword(secret.New(strings.Repeat("a", 73)))
		require.Error(t, err)
		require.ErrorIs(t, err, ErrPasswordTooLong)
		require.Equal(t, Password{}, passwd)
	})

	t.Run("Valid", func(t *testing.T) {
		passwd, err := NewPassword(secret.New("s3cureS3cret"))
		require.NoError(t, err)
		require.Equal(t, passwd.ExposeSecret(), "s3cureS3cret")
	})
}
