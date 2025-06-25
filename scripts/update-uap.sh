#!/usr/bin/env bash

set -uo pipefail

repository_root="$(git rev-parse --show-toplevel)"

echo "downloading latest version of $repository_root/pkg/embedded/uap/regexes.yml"
curl https://raw.githubusercontent.com/ua-parser/uap-core/master/regexes.yaml \
	-o "$repository_root/pkg/embedded/uap/regexes.yml"

echo "applying patch to $repository_root/pkg/embedded/uap/regexes.yml"
git apply --reject "$repository_root/pkg/embedded/uap/regexes.patch"

if [ "$?" = "0" ]; then
	$repository_root/scripts/update-uapatch.sh
else
	echo "failed to apply one or more hunks, fix it/them before running $repository_root/scripts/update-uapatch.sh to update patch"
fi
