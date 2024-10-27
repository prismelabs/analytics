#!/usr/bin/env bash

set -euo pipefail

repository_root="$(git rev-parse --show-toplevel)"

curl https://raw.githubusercontent.com/ua-parser/uap-core/master/regexes.yaml \
	-o "$repository_root/pkg/embedded/uap/regexes.yml"

git apply --check "$repository_root/pkg/embedded/uap/regexes.patch"

git apply "$repository_root/pkg/embedded/uap/regexes.patch"

