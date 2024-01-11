package secretstring

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecretString(t *testing.T) {
	t.Run("PrintSecretString", func(t *testing.T) {
		buf := bytes.Buffer{}
		secret := NewSecretString("MyUltraSecretString")

		fmt.Fprintf(&buf, "%v", secret)
		require.Equal(t, "SecretString(******)", buf.String())

		buf.Reset()

		fmt.Fprintf(&buf, "%+v", secret)
		require.Equal(t, "SecretString(******)", buf.String())
	})
}
