package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/negrel/configue"
	"github.com/prismelabs/analytics/pkg/embedded"
)

func grafanaDashboard() {
	type Params struct {
		DatasourceClickhouseUid string
	}
	var params Params

	figue := configue.New(
		"grafana-dashboard",
		configue.ContinueOnError,
		configue.NewFlag(),
	)
	figue.Usage = func() {
		_, _ = fmt.Fprintln(figue.Output(), "prisme - High-perfomance, self-hosted and privacy-focused web analytics service.")
		_, _ = fmt.Fprintln(figue.Output())
		_, _ = fmt.Fprintln(figue.Output(), "Usage:")
		_, _ = fmt.Fprintln(figue.Output(), "  prisme grafana-dashboard [FLAGS]")
		_, _ = fmt.Fprintln(figue.Output())
		_, _ = fmt.Fprintln(figue.Output(), "  prisme grafana-dashboard -datasource-clickhouse-uid a2fadcc3-80e5-451e-86cf-70ca584bf626")
		_, _ = fmt.Fprintln(figue.Output())
		figue.PrintDefaults()
	}
	figue.StringVar(&params.DatasourceClickhouseUid, "datasource.clickhouse.uid",
		"", "identifier of ClickHouse datasource in Grafana")

	err := figue.Parse()
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		cliError(err)
	}
	if params.DatasourceClickhouseUid == "" {
		cliError(fmt.Errorf("invalid -datasource-clickhouse-uid"))
	}

	tpl := template.Must(template.New("grafana-dashboards").
		ParseFS(embedded.GrafanaDashboards, "grafana_dashboards/*"))

	err = tpl.ExecuteTemplate(os.Stdout, "web_analytics.json", params)
	if err != nil {
		panic(err)
	}
}
