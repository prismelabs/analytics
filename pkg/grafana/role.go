package grafana

import (
	"encoding/json"
	"fmt"
)

// Role enumerate possible user roles in an organization.
// If role is unknown, it will default to None.
type Role int8

const (
	RoleNone Role = iota
	RoleViewer
	RoleEditor
	RoleAdmin
)

// String implements fmt.Stringer.
func (r Role) String() string {
	switch r {
	case RoleViewer:
		return "Viewer"
	case RoleEditor:
		return "Editor"
	case RoleAdmin:
		return "Admin"
	case RoleNone:
		return "None"
	default:
		panic(fmt.Errorf("unknown role: %v", int8(r)))
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (r *Role) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case `"Viewer"`:
		*r = RoleViewer
	case `"Editor"`:
		*r = RoleEditor
	case `"Admin"`:
		*r = RoleAdmin
	case `"None"`:
		*r = RoleNone
	default:
		return fmt.Errorf("unknown role: %v", string(data))
	}

	return nil
}

// MarshalJSON implements json.Marshaler.
func (r Role) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}
