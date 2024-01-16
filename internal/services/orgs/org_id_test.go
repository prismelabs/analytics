package orgs

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestParseUserId(t *testing.T) {
	t.Run("NotAUuid", func(t *testing.T) {
		uid, err := ParseOrgId("")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrOrgIdIsNotAUuidV4)
		require.Equal(t, OrgId{}, uid)
	})

	t.Run("UuidV6", func(t *testing.T) {
		uid, err := ParseOrgId(uuid.Must(uuid.NewV6()).String())
		require.Error(t, err)
		require.ErrorIs(t, err, ErrOrgIdIsNotAUuidV4)
		require.Equal(t, OrgId{}, uid)
	})

	t.Run("UuidV4", func(t *testing.T) {
		rawUuid := uuid.New().String()
		uid, err := ParseOrgId(rawUuid)
		require.NoError(t, err)
		require.Equal(t, rawUuid, uid.String())
	})
}
