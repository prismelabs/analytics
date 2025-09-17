package dataview

import (
	"encoding/json/jsontext"
	"errors"
	"io"
	"strings"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/valyala/fasthttp"
)

var (
	ErrInvalidData   = errors.New("invalid data")
	ErrNotJsonObject = errors.New("not a JSON object")
)

// KvCollector define objects that can collect a set of keys and values.
type KvCollector interface {
	CollectKeysValues() (keys, values []string, err error)
}

// NewJsonKvCollector returns a new JsonKvCollector that will read a JSON object
// from provided io.Reader.
func NewJsonKvCollector(r io.Reader) JsonKvCollector {
	return JsonKvCollector{decoder: jsontext.NewDecoder(r)}
}

// JsonKvCollector collects keys and values from a JSON object.
// Values are raw JSON encoded value: string values will contains quotes.
type JsonKvCollector struct {
	decoder *jsontext.Decoder
}

// CollectKeysValues implements KvCollector.
// JsonKvCollector collects values as raw JSON encoded string: string value will
// contains quotes.
func (jkc JsonKvCollector) CollectKeysValues() (keys, values []string, err error) {
	switch jkc.decoder.PeekKind() {
	case '{':
		_, err := jkc.decoder.ReadToken()
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, ErrNotJsonObject
	}

	for jkc.decoder.PeekKind() != '}' {
		var key, val jsontext.Value
		key, err = jkc.decoder.ReadValue()
		if err != nil {
			return
		}
		keys = append(keys, unquote(key.String()))

		val, err = jkc.decoder.ReadValue()
		if err != nil {
			return
		}
		values = append(values, val.String())
	}

	return
}

func unquote(str string) string {
	return str[1 : len(str)-1]
}

// FasthttpArgsKeysValuesCollector implements KvCollector.
// FasthttpArgsKeysValuesCollector collects values whose key have given prefix
// as raw string.
type FasthttpArgsKeysValuesCollector struct {
	Args           *fasthttp.Args
	Prefix         string
	ValueValidator func([]byte) bool
}

// CollectKeysValues implements keysValuesCollector.
func (fakvc FasthttpArgsKeysValuesCollector) CollectKeysValues() (keys, values []string, err error) {
	fakvc.Args.VisitAll(func(keyBytes, valueBytes []byte) {
		if err == nil && len(keyBytes) > len(fakvc.Prefix) &&
			strings.HasPrefix(utils.UnsafeString(keyBytes), fakvc.Prefix) {

			// Validate arg.
			if valid := fakvc.ValueValidator(valueBytes); !valid {
				err = ErrInvalidData
				keys = nil
				values = nil
				return
			}

			// Copy key and value.
			keys = append(keys, string(keyBytes[len(fakvc.Prefix):]))
			values = append(values, string(valueBytes))
		}
	})

	return
}
