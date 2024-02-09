package grafana

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"text/template"

	"github.com/prismelabs/prismeanalytics/pkg/config"
	"github.com/prismelabs/prismeanalytics/pkg/embedded"
	"github.com/prismelabs/prismeanalytics/pkg/grafana"
)

// Service define a Grafana ressources management service.
type Service interface {
	SetupDatasourceAndDashboards(context.Context, grafana.OrgId) error
}

// ProvideService is a wire provider for grafana service.
func ProvideService(cli grafana.Client, cfg config.Clickhouse) Service {
	tmpl, err := template.ParseFS(embedded.GrafanaDashboards, "grafana_dashboards/*")
	if err != nil {
		panic(fmt.Errorf("failed to parse grafana dashboards template: %w", err))
	}

	return service{
		embedded.GrafanaDashboards,
		tmpl,
		cli,
		cfg,
	}
}

type service struct {
	staticDashboards fs.FS
	tmpl             *template.Template
	cli              grafana.Client
	chCfg            config.Clickhouse
}

// SetupDatasourceAndDashboards implements Service.
func (s service) SetupDatasourceAndDashboards(ctx context.Context, orgId grafana.OrgId) error {
	// Retrieve datasource.
	ds, err := s.cli.GetDatasourceByName(ctx, orgId, "Prisme Analytics")
	if errors.Is(err, grafana.ErrGrafanaDatasourceNotFound) {
		// Create it if needed.
		ds, err = s.cli.CreateDatasource(ctx, orgId, "Prisme Analytics", "grafana-clickhouse-datasource", false)
		if err != nil {
			return fmt.Errorf("failed create clickhouse datasource in grafana: %w", err)
		}

		host, port, err := net.SplitHostPort(s.chCfg.HostPort)
		if err != nil {
			panic(fmt.Errorf("failed to split clickhouse host port: %w", err))
		}

		ds.JSONData = map[string]any{
			"server":          host,
			"port":            port,
			"username":        s.chCfg.User.ExposeSecret(),
			"validate":        true,
			"defaultDatabase": s.chCfg.Database,
			"protocol":        "native",
			"secure":          s.chCfg.TlsEnabled,
		}
		ds.SecureJSONData = map[string]any{
			"password": s.chCfg.Password.ExposeSecret(),
		}

		err = s.cli.UpdateDatasource(ctx, orgId, ds)
		if err != nil {
			return fmt.Errorf("failed to update clickhouse datasource in grafana: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("failed create/get clickhouse datasource in grafana: %w", err)
	}
	dsId := ds.UID

	// Generate dahsboard.json from template.
	var generalDashboardJSON map[string]any
	{
		type GeneralTemplateData struct {
			DatasourceClickhouseUID grafana.DatasourceID
		}

		buf := bytes.Buffer{}
		err := s.tmpl.ExecuteTemplate(&buf, "general.json", GeneralTemplateData{
			DatasourceClickhouseUID: dsId,
		})
		if err != nil {
			panic(fmt.Sprintf("failed to execute dashboard template: %v", err.Error()))
		}

		// Unmarshal json.
		err = json.Unmarshal(buf.Bytes(), &generalDashboardJSON)
		if err != nil {
			panic(err)
		}
	}

	// Load static home.json dashboard.
	var homeDashboardJSON map[string]any
	{
		// Load Home dashboard JSON.
		homeDashboardFile, err := s.staticDashboards.Open("grafana_dashboards/home.json")
		if err != nil {
			panic(err)
		}
		homeDashboardBuf, err := io.ReadAll(homeDashboardFile)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(homeDashboardBuf, &homeDashboardJSON)
		if err != nil {
			panic(err)
		}
	}

	// Setup Built in folder.
	var folderId grafana.FolderID
	{
		folderName := "Prisme Analytics"

		// Create folder.
		folder, err := s.cli.CreateFolder(ctx, orgId, folderName)
		if err != nil {
			if err == grafana.ErrGrafanaFolderAlreadyExists {
				return nil
			}
			return fmt.Errorf("failed to create %q grafana folder: %w", folderName, err)
		}
		folderId = folder.Uid

		// Remove default permissions.
		err = s.cli.SetFolderPermissions(
			ctx,
			orgId,
			folder.Uid,
			grafana.FolderPermission{
				Permission: grafana.FolderPermissionLevelView,
				Role:       grafana.RoleEditor,
			},
			grafana.FolderPermission{
				Permission: grafana.FolderPermissionLevelView,
				Role:       grafana.RoleViewer,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to set %q grafana folder permissions: %w", folderName, err)
		}
	}

	// Setup general dashboard.
	{
		_, err := s.cli.CreateUpdateDashboard(ctx, orgId, folderId, generalDashboardJSON, true)
		if err != nil {
			return fmt.Errorf("failed to create/update general grafana dashboard: %w", err)
		}
	}

	// Setup home dashboard.
	{
		_, err := s.cli.CreateUpdateDashboard(ctx, orgId, folderId, homeDashboardJSON, true)
		if err != nil {
			return fmt.Errorf("failed to create/update built-in home grafana dashboard: %w", err)
		}
	}

	return nil
}