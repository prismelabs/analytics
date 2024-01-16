package orgs

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/rivo/uniseg"
)

var (
	ErrOrgNameTooShort = errors.New("org name too short")
	ErrOrgNameTooLong  = errors.New("org name too long")
)

// OrgName define an organization name.
// An organization name is considered valid if it contains at least 3 non whitespace
// character.
type OrgName struct {
	value string
}

// NewOrgName returns a new org name from the given string.
// An error is returned if the given value doesn't satisfy org name requirements.
func NewOrgName(value string) (OrgName, error) {
	gcCount := uniseg.GraphemeClusterCount(value)
	if gcCount < 3 {
		return OrgName{}, ErrOrgNameTooShort
	} else if gcCount >= 64 {
		return OrgName{}, ErrOrgNameTooLong
	}

	return OrgName{value}, nil
}

// String implements fmt.Stringer.
func (on OrgName) String() string {
	return on.value
}

// Scan implements sql.Scanner.
func (on *OrgName) Scan(src any) error {
	if t, ok := src.(string); ok {
		on.value = t
		return nil
	}
	return fmt.Errorf("cannot scan %T into OrgName", src)
}

// Value implements driver.Valuer.
func (on OrgName) Value() (driver.Value, error) {
	return on.String(), nil
}
