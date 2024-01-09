#!/usr/bin/env bash

set -euo pipefail

setenv() {
	printf "export %s='%s'\n" "$1" "$2"
}

# Server options.
setenv PRISME_ACCESS_LOG "/dev/stdout"
setenv PRISME_DEBUG "true"
setenv PRISME_PORT "8000"
setenv PRISME_TRUST_PROXY "false"

# Postgres related options.
setenv PRISME_POSTGRES_URL "postgres://postgres:password@postgres.localhost:5432/prisme?sslmode=disable"
