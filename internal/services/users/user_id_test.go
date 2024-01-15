package users

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestParseUserId(t *testing.T) {
	t.Run("NotAUuid", func(t *testing.T) {
		uid, err := ParseUserId("")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUserIdIsNotAUuidV4)
		require.Equal(t, UserId{}, uid)
	})

	t.Run("UuidV6", func(t *testing.T) {
		uid, err := ParseUserId(uuid.Must(uuid.NewV6()).String())
		require.Error(t, err)
		require.ErrorIs(t, err, ErrUserIdIsNotAUuidV4)
		require.Equal(t, UserId{}, uid)
	})

	t.Run("UuidV4", func(t *testing.T) {
		rawUuid := uuid.New().String()
		uid, err := ParseUserId(rawUuid)
		require.NoError(t, err)
		require.Equal(t, rawUuid, uid.String())
	})
}
