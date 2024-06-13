package grafana

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

const MaxUIDLength = 40

var (
	ErrUIDTooLong       = fmt.Errorf("UID is longer than %d symbols", MaxUIDLength)
	ErrUIDFormatInvalid = errors.New("invalid format of UID. Only letters, numbers, '-' and '_' are allowed")
	ErrUIDEmpty         = fmt.Errorf("UID is empty")

	validUIDCharPattern = `a-zA-Z0-9\-\_`
	validUIDPattern     = regexp.MustCompile(`^[` + validUIDCharPattern + `]*$`).MatchString
	isValidShortUID     = validUIDPattern
)

// Uid represents grafana uid type used to identify multiples ressources.
type Uid struct {
	value string
}

// ParseUid parses and validates given uid.
func ParseUid(uid string) (Uid, error) {
	if len(uid) == 0 {
		return Uid{}, ErrUIDEmpty
	}
	if len(uid) > MaxUIDLength {
		return Uid{}, ErrUIDTooLong
	}
	if !isValidShortUID(uid) {
		return Uid{}, ErrUIDFormatInvalid
	}

	return Uid{uid}, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (uid *Uid) UnmarshalJSON(rawJSON []byte) error {
	rawJSON = bytes.TrimPrefix(rawJSON, []byte(`"`))
	rawJSON = bytes.TrimSuffix(rawJSON, []byte(`"`))

	// Grafana may return empty UID.
	if len(rawJSON) == 0 {
		return nil
	}

	var err error
	*uid, err = ParseUid(string(rawJSON))
	if err != nil {
		return err
	}

	return nil
}

// MarshalJSON implements json.Marshaler.
func (uid Uid) MarshalJSON() ([]byte, error) {
	return json.Marshal(uid.value)
}

// String implements fmt.Stringer.
func (u Uid) String() string {
	return u.value
}
