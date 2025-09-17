package dataview

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestJsonKvCollector(t *testing.T) {
	t.Run("CollectKeysValues/Empty", func(t *testing.T) {
		kvCollector := NewJsonKvCollector(bytes.NewReader([]byte(`{}`)))
		keys, values, err := kvCollector.CollectKeysValues()
		require.NoError(t, err)

		require.Nil(t, keys)
		require.Nil(t, values)
	})
	t.Run("CollectKeysValues/NonEmpty", func(t *testing.T) {
		kvCollector := NewJsonKvCollector(bytes.NewReader([]byte(`{"foo":"bar","bool":true,"number":1.123,"null":null,"obj":{"foo":"bar","bool":true}}`)))
		keys, values, err := kvCollector.CollectKeysValues()
		require.NoError(t, err)

		require.Equal(t, []string{"foo", "bool", "number", "null", "obj"}, keys)
		require.Equal(t, []string{`"bar"`, "true", "1.123", "null", `{"foo":"bar","bool":true}`}, values)
	})

	t.Run("CollectKeysValues/MalformedJson", func(t *testing.T) {
		// Missing closing brace.
		kvCollector := NewJsonKvCollector(bytes.NewReader([]byte(
			`{"foo":"bar","bool":true,"number":1,"null":null,"obj":{"foo":"bar","bool":true}`,
		)))
		_, _, err := kvCollector.CollectKeysValues()
		require.Error(t, err)
	})
}

func TestFasthttpArgsKvCollector(t *testing.T) {
	t.Run("CollectKeysValues/NoPrefix", func(t *testing.T) {
		kvCollector := FasthttpArgsKeysValuesCollector{
			Args:           &fasthttp.Args{},
			ValueValidator: json.Valid,
		}
		kvCollector.Args.Add("foo", `"bar"`)
		kvCollector.Args.Add("number", "1")

		keys, values, err := kvCollector.CollectKeysValues()
		require.NoError(t, err)

		require.Equal(t, []string{"foo", "number"}, keys)
		require.Equal(t, []string{`"bar"`, "1"}, values)
	})

	t.Run("CollectKeysValues/FooPrefix", func(t *testing.T) {
		kvCollector := FasthttpArgsKeysValuesCollector{
			Args:           &fasthttp.Args{},
			Prefix:         "foo-",
			ValueValidator: json.Valid,
		}
		kvCollector.Args.Add("foo-foo", `"bar"`)
		kvCollector.Args.Add("foo-bar", "1")

		keys, values, err := kvCollector.CollectKeysValues()
		require.NoError(t, err)

		require.Equal(t, []string{"foo", "bar"}, keys)
		require.Equal(t, []string{`"bar"`, "1"}, values)
	})

	t.Run("CollectKeysValues/MalformedJson", func(t *testing.T) {
		kvCollector := FasthttpArgsKeysValuesCollector{
			Args:           &fasthttp.Args{},
			Prefix:         "foo-",
			ValueValidator: json.Valid,
		}
		kvCollector.Args.Add("foo-foo", `bar`) // Missing double quote for a string.
		kvCollector.Args.Add("foo-bar", "1")

		keys, values, err := kvCollector.CollectKeysValues()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidData)
		require.Nil(t, keys)
		require.Nil(t, values)
	})
}
