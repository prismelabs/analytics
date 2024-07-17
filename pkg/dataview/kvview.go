package dataview

import (
	"errors"

	"github.com/gofiber/fiber/v2/utils"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
)

var (
	ErrKvViewValueIsNotAString = errors.New("value is not a string")
	ErrKvViewEntryNotFound     = errors.New("entry not found")
)

// KvView is a view over a set of key values.
type KvView interface {
	GetString(string) (string, error)
}

// JsonKvView is a KvView implementation over a JSON object.
type JsonKvView struct {
	Json *JsonData
}

// GetString implements KvView.
func (jkv JsonKvView) GetString(key string) (string, error) {
	data, err := jkv.Json.Data()
	if err != nil {
		return "", err
	}

	result := gjson.GetBytes(data, key)
	if !result.Exists() {
		return "", ErrKvViewEntryNotFound
	}

	if result.Type != gjson.String {
		return "", ErrKvViewValueIsNotAString
	}

	return result.Str, nil
}

// FasthttpArgsKvView is KvView implementation based on *fasthttp.Args.
type FasthttpArgsKvView struct {
	Args *fasthttp.Args
}

func (fakv FasthttpArgsKvView) GetString(key string) (string, error) {
	if !fakv.Args.Has(key) {
		return "", ErrKvViewEntryNotFound
	}

	return utils.UnsafeString(fakv.Args.Peek(key)), nil
}
