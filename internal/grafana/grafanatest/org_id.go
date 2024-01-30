package grafanatest

import (
	"math/rand"

	"github.com/prismelabs/prismeanalytics/internal/grafana"
)

func NewGrafanaOrgID() grafana.OrgId {
	return grafana.OrgId(rand.Int63())
}
