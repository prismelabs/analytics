package dataview

import (
	"errors"

	"github.com/tidwall/gjson"
)

var (
	ErrMalformedData = errors.New("malformed data")
)

// Data define a slice of byte encoded.
type Data []byte

// JsonData define data that is potentially encoded as JSON value.
type JsonData struct {
	data      Data
	validated bool
}

// NewJsonData creates a new JsonData holding given untrusted inner data.
func NewJsonData(data Data) *JsonData {
	return &JsonData{
		data:      data,
		validated: false,
	}
}

// Data returns underlying data and a nil error if it contains a well formed
// JSON.
func (jd *JsonData) Data() (Data, error) {
	if !jd.validated {
		if gjson.ValidBytes(jd.data) {
			jd.validated = true
		} else {
			return nil, ErrMalformedData
		}
	}

	return jd.data, nil
}

// JsonValidator returns nil if input is valid json and ErrMalformedData
// otherwise.
func JsonValidator(v []byte) error {
	if gjson.ValidBytes(v) {
		return nil
	}

	return ErrMalformedData
}
