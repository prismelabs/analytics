package event

import (
	"errors"

	"github.com/valyala/fasthttp"
)

var (
	slashPath        = []byte("/")
	ErrUriIsRelative = errors.New("uri is relative")
)

// Uri wraps a fasthttp.URI to implements Uri.
type Uri struct {
	fasthttp.URI
}

// IsEmpty returns whether Uri contains actual data or is empty.
func (u *Uri) IsEmpty() bool {
	return len(u.Host()) == 0
}

// Path returns URI path, i.e. /foo/bar of [http://aaa.com/foo/bar?baz=123#qwe](http://aaa.com/foo/bar?baz=123#qwe) .
//
// The returned path is always urldecoded and normalized, i.e. '//f%20obar/baz/../zzz' becomes '/f obar/zzz'.
//
// The returned bytes are valid until the next URI method call.
func (u *Uri) Path() []byte {
	p := u.URI.Path()
	if len(p) == 0 {
		return slashPath
	}

	if len(p) > 1 && p[len(p)-1] == '/' {
		return p[:len(p)-1]
	}

	return p
}

// Parse initializes URI from the given absolute uri.
func (u *Uri) Parse(uri []byte) error {
	err := u.URI.Parse(nil, uri)
	if err != nil {
		return err
	}

	if len(u.Host()) == 0 {
		return ErrUriIsRelative
	}

	return nil
}

// ReferrerUri wraps an URI to represent referrer URIs.
// An empty referrer uri is considered as "direct".
type ReferrerUri struct {
	Uri
}

// Parse initializes URI from the given absolute uri.
// If given uri is nil or if this method is never called, ReferrerUri is considered
// as direct.
func (ru *ReferrerUri) Parse(uri []byte) error {
	if len(uri) == 0 {
		return nil
	}

	return ru.Uri.Parse(uri)
}

// HostOrDirect returns uri host or "direct" if uri is empty.
func (ru *ReferrerUri) HostOrDirect() []byte {
	if ru.IsEmpty() {
		return []byte("direct")
	}

	return ru.Host()
}
