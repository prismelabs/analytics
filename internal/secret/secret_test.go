package secret

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecretString(t *testing.T) {
	t.Run("PrintSecretString", func(t *testing.T) {
		buf := bytes.Buffer{}
		secret := New("MyUltraSecretString")

		fmt.Fprintf(&buf, "%v", secret)
		require.Equal(t, "Secret[string](******)", buf.String())

		buf.Reset()

		fmt.Fprintf(&buf, "%+v", secret)
		require.Equal(t, "Secret[string](******)", buf.String())
	})
}
