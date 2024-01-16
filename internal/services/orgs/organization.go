package orgs

import (
	"time"

	"github.com/prismelabs/prismeanalytics/internal/services/users"
)

type Organization struct {
	Id           OrgId
	OwnerId      users.UserId
	Name         OrgName
	GrafanaOrgId GrafanaOrgId
	CreatedAt    time.Time
}
