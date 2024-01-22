#!/usr/bin/env bash

if [ $# -eq 0 ]; then
	echo "missing VERSION argument" >&2
	exit 1
fi

# If doesn't match semantic versioning, return input as is.
if ! egrep "^v[0-9]+.[0-9].[0-9]$" &> /dev/null <<< "$1"; then
	echo $1
	exit 0
fi

IFS=. read -a tags <<< "$1"

for (( i=1; i<${#tags[@]}; i++ )); do
	j=$((i-1))
	tags[i]=${tags[j]}.${tags[i]}
done

echo "latest" ${tags[@]}

