package embedded

import "embed"

//go:embed grafana_dashboards
var GrafanaDashboards embed.FS
