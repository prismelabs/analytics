package uri

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUri(t *testing.T) {
	t.Run("Parse", func(t *testing.T) {
		type testCase struct {
			uri                             string
			scheme, host, path, hash, query string
			expectedError                   error
		}

		testCases := []testCase{
			{
				uri:    "https://example.org:8080/foo/bar?q=baz#bang",
				scheme: "https",
				host:   "example.org:8080",
				path:   "/foo/bar",
				query:  "q=baz",
				hash:   "bang",
			},
			{
				uri:    "https://example.org:8080/foo/../bar#bang?foo=bar",
				scheme: "https",
				host:   "example.org:8080",
				path:   "/bar",
				query:  "",
				hash:   "bang?foo=bar",
			},
			{
				uri:    "https://example.org:8080",
				scheme: "https",
				host:   "example.org:8080",
				path:   "/",
				query:  "",
				hash:   "",
			},
			{
				uri:           "./hello/world",
				expectedError: ErrUriIsRelative,
			},
		}

		for _, tcase := range testCases {
			t.Run(tcase.uri, func(t *testing.T) {
				uri, err := Parse(tcase.uri)

				if tcase.expectedError == nil {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
					require.ErrorIs(t, err, tcase.expectedError)
					return
				}

				require.Equal(t, tcase.scheme, uri.Scheme())
				require.Equal(t, tcase.host, uri.Host())
				require.Equal(t, tcase.path, uri.Path())
				require.Equal(t, tcase.query, uri.QueryString())
				require.Equal(t, tcase.hash, uri.Hash())
			})
		}

		t.Run("ParsedUriIsCopied", func(t *testing.T) {
			rawUri := []byte("https://www.example.com/")

			uri, err := ParseBytes(rawUri)
			require.NoError(t, err)
			require.Equal(t, "www.example.com", uri.Host())

			// Edit rawUri buffer.
			rawUri[len("https://")+1] = 'x'

			require.Equal(t, "www.example.com", uri.Host())
		})
	})

	t.Run("String", func(t *testing.T) {
		t.Run("WithQueryArgs", func(t *testing.T) {
			uri, err := Parse("https://www.example.com?q=foo#bar")
			require.NoError(t, err)

			require.Equal(t, "https://www.example.com/?q=foo#bar", uri.String())
		})
	})
}