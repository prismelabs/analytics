package dataview

import (
	"errors"

	"github.com/tidwall/gjson"
)

var (
	ErrMalformedData = errors.New("malformed data")
)

// dataState holds whether data is unknown, was validated and is (or not) valid.
type dataState uint8

const (
	unknownData dataState = iota
	invalidData
	validData
)

// Data define a slice of byte encoded.
type Data []byte

// JsonData define data that is potentially encoded as JSON value.
type JsonData struct {
	data  Data
	state dataState
}

// NewJsonData creates a new JsonData holding given untrusted inner data.
func NewJsonData(data Data) *JsonData {
	return &JsonData{
		data:  data,
		state: unknownData,
	}
}

// Data returns underlying data and a nil error if it contains a well formed
// JSON.
func (jd *JsonData) Data() (Data, error) {
	if jd.state == unknownData {
		if gjson.ValidBytes(jd.data) {
			jd.state = validData
		} else {
			jd.state = invalidData
		}
	}

	switch jd.state {
	case invalidData:
		return nil, ErrMalformedData

	case validData:
		return jd.data, nil

	case unknownData:
		fallthrough
	default:
		panic("unexpected dataview.dataState")
	}
}

// JsonValidator returns nil if input is valid json and ErrMalformedData
// otherwise.
func JsonValidator(v []byte) error {
	if gjson.ValidBytes(v) {
		return nil
	}

	return ErrMalformedData
}
