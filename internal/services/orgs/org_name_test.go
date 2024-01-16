package orgs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewOrgName(t *testing.T) {
	t.Run("TooShort", func(t *testing.T) {
		un, err := NewOrgName("🏳️‍🌈")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrOrgNameTooShort)
		require.Equal(t, OrgName{}, un)
	})

	t.Run("TooLong", func(t *testing.T) {
		un, err := NewOrgName(strings.Repeat("Foo ", 16)) // length of 64
		require.Error(t, err)
		require.ErrorIs(t, err, ErrOrgNameTooLong)
		require.Equal(t, OrgName{}, un)
	})

	t.Run("Valid", func(t *testing.T) {
		un, err := NewOrgName("Foo")
		require.NoError(t, err)
		require.Equal(t, "Foo", un.String())
	})
}
