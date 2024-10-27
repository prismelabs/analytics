#!/usr/bin/env bash

set -euo pipefail

repository_root="$(git rev-parse --show-toplevel)"

gengeommdb > "$repository_root/pkg/embedded/geodb/ip2asn-combined.mmdb"

