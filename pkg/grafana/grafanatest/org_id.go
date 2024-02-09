package grafanatest

import (
	"math/rand"

	"github.com/prismelabs/analytics/pkg/grafana"
)

func NewGrafanaOrgID() grafana.OrgId {
	return grafana.OrgId(rand.Int63())
}
