package dataview

import (
	"strings"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
)

// KvCollector define objects that can collect a set of keys and values.
type KvCollector interface {
	CollectKeysValues() (keys, values []string, err error)
}

// JsonKvCollector collects keys and values from a JSON object.
// Values are raw JSON encoded value: string values will contains quotes.
type JsonKvCollector struct {
	Json *JsonData
	Path string
}

// CollectKeysValues implements KvCollector.
// JsonKvCollector collects values as raw JSON encoded string: string value will
// contains quotes.
func (jkc JsonKvCollector) CollectKeysValues() (keys, values []string, err error) {
	data, err := jkc.Json.Data()
	if err != nil {
		return nil, nil, err
	}

	// Get keys.
	result := gjson.GetBytes(data, jkc.Path+"@keys")
	result.ForEach(func(_, key gjson.Result) bool {
		keys = append(keys, utils.CopyString(key.String()))
		return true
	})

	// Get values.
	result = gjson.GetBytes(data, jkc.Path+"@values")
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
	Args           *fasthttp.Args
	Prefix         string
	ValueValidator func([]byte) error
}

// CollectKeysValues implements keysValuesCollector.
func (fakvc FasthttpArgsKeysValuesCollector) CollectKeysValues() (keys, values []string, err error) {
	fakvc.Args.VisitAll(func(keyBytes, valueBytes []byte) {
		if err == nil && len(keyBytes) > len(fakvc.Prefix) &&
			strings.HasPrefix(utils.UnsafeString(keyBytes), fakvc.Prefix) {

			// Validate arg.
			if valueErr := fakvc.ValueValidator(valueBytes); valueErr != nil {
				err = valueErr
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
