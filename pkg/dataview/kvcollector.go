package dataview

import (
	"strings"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
)

// KvCollector define objects that can collect a set of keys and values.
type KvCollector interface {
	CollectKeysValues() (keys, values []string)
}

// JsonKvCollector collects keys and values from a JSON object.
// Values are raw JSON encoded value: string values will contains quotes.
type JsonKvCollector struct {
	Json []byte
	Path string
}

// CollectKeysValues implements KvCollector.
// JsonKvCollector collects values as raw JSON encoded string: string value will
// contains quotes.
func (jkc JsonKvCollector) CollectKeysValues() (keys, values []string) {
	// Get keys.
	result := gjson.GetBytes(jkc.Json, jkc.Path+"@keys")
	result.ForEach(func(_, key gjson.Result) bool {
		keys = append(keys, utils.CopyString(key.String()))
		return true
	})

	// Get values.
	result = gjson.GetBytes(jkc.Json, jkc.Path+"@values")
	result.ForEach(func(_, value gjson.Result) bool {
		values = append(values, utils.CopyString(value.Raw))
		return true
	})

	return
}

// FasthttpArgsKeysValuesCollector implements KvCollector.
// FasthttpArgsKeysValuesCollector collects values whose key have given prefix
// as raw string.
type FasthttpArgsKeysValuesCollector struct {
	Args   *fasthttp.Args
	Prefix string
}

// collectKeysValues implements keysValuesCollector.
func (fakvc FasthttpArgsKeysValuesCollector) CollectKeysValues() (keys, values []string) {
	fakvc.Args.VisitAll(func(keyBytes, valueBytes []byte) {
		if len(keyBytes) > len(fakvc.Prefix) &&
			strings.HasPrefix(utils.UnsafeString(keyBytes), fakvc.Prefix) {
			// Copy key and value.
			keys = append(keys, string(keyBytes[len(fakvc.Prefix):]))
			values = append(values, string(valueBytes))
		}
	})

	return
}
