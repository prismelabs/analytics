package dataview

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestJsonKvCollector(t *testing.T) {
	t.Run("CollectKeysValues/NoPath", func(t *testing.T) {
		kvCollector := JsonKvCollector{Json: []byte(`{"foo":"bar","bool":true,"number":1,"null":null,"obj":{"foo":"bar","bool":true}}`)}
		keys, values := kvCollector.CollectKeysValues()

		require.Equal(t, []string{"foo", "bool", "number", "null", "obj"}, keys)
		require.Equal(t, []string{`"bar"`, "true", "1", "null", `{"foo":"bar","bool":true}`}, values)
	})

	t.Run("CollectKeysValues/WithPath", func(t *testing.T) {
		kvCollector := JsonKvCollector{Json: []byte(`{"foo":"bar","bool":true,"number":1,"null":null,"obj":{"foo":"bar","bool":true}}`), Path: "obj."}
		keys, values := kvCollector.CollectKeysValues()

		require.Equal(t, []string{"foo", "bool"}, keys)
		require.Equal(t, []string{`"bar"`, "true"}, values)
	})
}

func TestFasthttpArgsKvCollector(t *testing.T) {
	t.Run("CollectKeysValues/NoPrefix", func(t *testing.T) {
		kvCollector := FasthttpArgsKeysValuesCollector{Args: &fasthttp.Args{}}
		kvCollector.Args.Add("foo", "bar")
		kvCollector.Args.Add("number", "1")

		keys, values := kvCollector.CollectKeysValues()
		require.Equal(t, []string{"foo", "number"}, keys)
		require.Equal(t, []string{"bar", "1"}, values)
	})

	t.Run("CollectKeysValues/FooPrefix", func(t *testing.T) {
		kvCollector := FasthttpArgsKeysValuesCollector{Args: &fasthttp.Args{}, Prefix: "foo-"}
		kvCollector.Args.Add("foo-foo", "bar")
		kvCollector.Args.Add("foo-bar", "1")

		keys, values := kvCollector.CollectKeysValues()
		require.Equal(t, []string{"foo", "bar"}, keys)
		require.Equal(t, []string{"bar", "1"}, values)
	})
}
