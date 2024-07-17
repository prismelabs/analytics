package dataview

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestJsonKvView(t *testing.T) {
	t.Run("GetString", func(t *testing.T) {
		t.Run("EntryNotFound", func(t *testing.T) {
			kvView := JsonKvView{Json: NewJsonData([]byte(`{"foo":"bar"}`))}

			entry, err := kvView.GetString("baz")
			require.Error(t, err)
			require.ErrorIs(t, err, ErrKvViewEntryNotFound)
			require.Equal(t, "", entry)
		})

		t.Run("EntryNotAString", func(t *testing.T) {
			kvView := JsonKvView{Json: NewJsonData([]byte(`{"foo":1}`))}

			entry, err := kvView.GetString("foo")
			require.Error(t, err)
			require.ErrorIs(t, err, ErrKvViewValueIsNotAString)
			require.Equal(t, "", entry)
		})

		t.Run("Valid", func(t *testing.T) {
			kvView := JsonKvView{Json: NewJsonData([]byte(`{"foo":"bar","baz":"qux"}`))}

			entry, err := kvView.GetString("baz")
			require.NoError(t, err)
			require.Equal(t, "qux", entry)
		})
	})
}

func TestFasthttpArgsKvView(t *testing.T) {
	t.Run("GetString", func(t *testing.T) {
		t.Run("EntryNotFound", func(t *testing.T) {
			kvView := FasthttpArgsKvView{Args: &fasthttp.Args{}}
			kvView.Args.Add("foo", "bar")

			entry, err := kvView.GetString("baz")
			require.Error(t, err)
			require.ErrorIs(t, err, ErrKvViewEntryNotFound)
			require.Equal(t, "", entry)
		})

		t.Run("Valid", func(t *testing.T) {
			kvView := FasthttpArgsKvView{Args: &fasthttp.Args{}}
			kvView.Args.Add("foo", "bar")
			kvView.Args.Add("baz", "qux")

			entry, err := kvView.GetString("baz")
			require.NoError(t, err)
			require.Equal(t, "qux", entry)
		})
	})
}
