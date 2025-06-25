#!/usr/bin/env bash

set -uo pipefail

repository_root="$(git rev-parse --show-toplevel)"

echo "updating patch $repository_root/pkg/embedded/uap/regexes.patch"
git diff <(curl https://raw.githubusercontent.com/ua-parser/uap-core/master/regexes.yaml) pkg/embedded/uap/regexes.yml
