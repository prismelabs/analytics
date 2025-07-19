package testutils

import (
	"testing"

	"github.com/negrel/configue"
	"github.com/stretchr/testify/require"
)

// ConfigueLoadFunc creates a *configue.Figue and loads configuration options
// defined in provided function.
func ConfigueLoadFunc(t *testing.T, setup func(figue *configue.Figue)) {
	figue := configue.New(
		"",
		configue.ContinueOnError,
		configue.NewEnv("PRISME"),
	)
	setup(figue)
	err := figue.Parse()
	require.NoError(t, err, "failed to parse configuration options")
}

// ConfigueLoad loads all objects provided.
func ConfigueLoad(t *testing.T, objects ...interface{ RegisterOptions(*configue.Figue) }) {
	ConfigueLoadFunc(t, func(figue *configue.Figue) {
		for _, obj := range objects {
			obj.RegisterOptions(figue)
		}
	})
}
