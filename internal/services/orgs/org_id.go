package orgs

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrOrgIdIsNotAUuidV4 = errors.New("org id is not a uuid version 4")
)

// OrgId define a unique organization identifier.
// Every UUIDv4 is a valid org id.
type OrgId struct {
	value uuid.UUID
}

// NewOrgId generates a new random org id.
func NewOrgId() OrgId {
	return OrgId{uuid.New()}
}

// ParseOrgId parses the given value as a org id.
// An error is returned if the id doesn't satisify org id requirements.
func ParseOrgId(value string) (OrgId, error) {
	uid, err := uuid.Parse(value)
	if err != nil {
		return OrgId{}, fmt.Errorf("%w: %w", ErrOrgIdIsNotAUuidV4, err)
	}
	if uid.Version() != 4 {
		return OrgId{}, ErrOrgIdIsNotAUuidV4
	}

	return OrgId{uid}, nil
}

// String implements fmt.Stringer.
func (oid OrgId) String() string {
	return oid.value.String()
}

// Scan implements sql.Scanner.
func (oid *OrgId) Scan(src any) error {
	nid := uuid.NullUUID{}
	err := nid.Scan(src)
	if err != nil {
		return err
	}
	if !nid.Valid {
		return fmt.Errorf("org ID can't be NULL")
	}
	if nid.UUID.Version() != 4 {
		return fmt.Errorf("invalid org id, UUID v4 expected, got %d", nid.UUID.Version())
	}

	copy(oid.value[:], nid.UUID[:])

	return nil
}

// Value implements driver.Valuer.
func (oid OrgId) Value() (driver.Value, error) {
	return oid.String(), nil
}
