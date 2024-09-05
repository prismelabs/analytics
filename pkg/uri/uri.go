package uri

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/valyala/fasthttp"
)

var ErrUriIsRelative = errors.New("uri is relative")

// Parse parses an URI and normalize it. Raw query args are copied and stored
// if extractQuery is true.
func Parse(uriStr string) (Uri, error) {
	return ParseBytes(utils.UnsafeBytes(uriStr))
}

// ParseBytes parses an URI and normalize it. Raw query args are copied and stored
// if extractQuery is true.
func ParseBytes(uri []byte) (Uri, error) {
	furi := fasthttp.URI{}
	err := furi.Parse(nil, uri)
	if err != nil {
		return Uri{}, err
	}

	if len(furi.Host()) == 0 {
		return Uri{}, ErrUriIsRelative
	}

	data := utils.CopyBytes(furi.FullURI())

	return Uri{
		data:      data,
		schemeLen: len(furi.Scheme()),
		hostLen:   len(furi.Host()),
		pathLen:   len(furi.Path()),
		queryLen:  len(furi.QueryString()),
		hashLen:   len(furi.Hash()),
	}, nil
}

// Uri define a read-only absolute URI object.
type Uri struct {
	data      []byte
	schemeLen int
	hostLen   int
	pathLen   int
	queryLen  int
	hashLen   int
}

// IsValid returns true if data was successfully parsed.
func (u *Uri) IsValid() bool {
	return u != nil && len(u.data) != 0
}

// Scheme returns scheme of URI: https for
// https://www.example.com:8080/foo/bar?q=baz#bang
func (u *Uri) Scheme() string {
	return utils.UnsafeString(u.data[:u.schemeLen])
}

// Host returns host of URI: www.example.com:8080 for
// https://www.example.com:8080/foo/bar?q=baz#bang
func (u *Uri) Host() string {
	start := u.schemeLen + len("://")
	return utils.UnsafeString(u.data[start : start+u.hostLen])
}

// Path returns normalized path of URI: /foo/bar for
// https://www.example.com:8080/foo/bar?q=baz#bang
func (u *Uri) Path() string {
	start := u.schemeLen + len("://") + u.hostLen
	return utils.UnsafeString(u.data[start : start+u.pathLen])
}

// QueryBytes returns extracted query bytes (extractQuery must be true when
// calling Parse / ParseBytes): q=baz for
// https://www.example.com:8080/foo/bar?q=baz#bang
func (u *Uri) QueryString() string {
	start := u.schemeLen + len("://") + u.hostLen + u.pathLen
	if u.queryLen > 0 {
		start++ // Add 1 for ?
	}
	return utils.UnsafeString(u.data[start : start+u.queryLen])
}

// Hash returns hash of URI: bang for
// https://www.example.com:8080/foo/bar?q=baz#bang
func (u *Uri) Hash() string {
	start := u.schemeLen + len("://") + u.hostLen + u.pathLen + u.queryLen
	if u.queryLen > 0 {
		start++ // Add 1 for ?
	}
	if u.hashLen > 0 {
		start++ // Add 1 for #
	}
	return utils.UnsafeString(u.data[start : start+u.hashLen])
}

// String implements fmt.Stringer.
func (u Uri) String() string {
	if u.queryLen == 0 {
		return utils.UnsafeString(u.data)
	}

	uri := u.Scheme() + "://" + u.Host() + u.Path()
	if u.queryLen > 0 {
		uri += "?" + u.QueryString()
	}

	if u.hashLen > 0 {
		uri += "#" + u.Hash()
	}

	return uri
}

// MarshalJSON implements json.Marshaler.
func (u Uri) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *Uri) UnmarshalJSON(b []byte) error {
	str := ""
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	*u, err = Parse(str)
	return err
}

// Value implements driver.Valuer.
func (u *Uri) Value() (driver.Value, error) {
	return u.String(), nil
}

// Scan implements sql.Scanner.
func (u *Uri) Scan(src any) error {
	if str, ok := src.(string); ok {
		var err error
		*u, err = Parse(str)
		return err
	}

	return fmt.Errorf("cannot scan %T into %T", src, u)
}
