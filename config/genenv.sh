#!/usr/bin/env bash

set -euo pipefail

setenv() {
	printf "export %s='%s'\n" "$1" "$2"
}

# Server options.
# setenv PRISME_MODE
setenv PRISME_ACCESS_LOG "/dev/stdout"
setenv PRISME_DEBUG "true"
setenv PRISME_PORT "8000"
setenv PRISME_TRUST_PROXY "false"

# Clickhouse related options.
setenv PRISME_CLICKHOUSE_TLS "false"
setenv PRISME_CLICKHOUSE_HOSTPORT "clickhouse.localhost:9000"
setenv PRISME_CLICKHOUSE_DB "prisme"
setenv PRISME_CLICKHOUSE_USER "clickhouse"
setenv PRISME_CLICKHOUSE_PASSWORD "password"

# Source registry options.
setenv PRISME_SOURCE_REGISTRY_SOURCES "localhost,mywebsite.localhost,foo.mywebsite.localhost"

# Grafana related options.
setenv PRISME_GRAFANA_URL "http://grafana.localhost:3000"
setenv PRISME_GRAFANA_USER "admin"
setenv PRISME_GRAFANA_PASSWORD "admin"
