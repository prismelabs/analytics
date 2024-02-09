package grafana

import (
	"errors"
	"time"
)

var (
	ErrGrafanaUserNotFound = errors.New("grafana user not found")
)

// UserId define a grafana user id.
type UserId int64

// User define a grafana user.
type User struct {
	Id                             UserId    `json:"id"`
	CreatedAt                      time.Time `json:"createdAt"`
	UpdatedAt                      time.Time `json:"updatedAt"`
	IsDisabled                     bool      `json:"isDisabled"`
	IsGrafanaAdmin                 bool      `json:"isGrafanaAdmin"`
	IsGrafanaAdminExternallySynced bool      `json:"isGrafanaAdminExternallySynced"`
	Name                           string    `json:"name"`
	Email                          string    `json:"email"`
}
