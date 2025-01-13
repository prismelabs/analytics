# Maintainer documentation

This document explains how to maintain the repository.

## Upgrade dependencies

It is important to regularly update direct and transitive dependencies to
mitigate security vulnerabilities.

First, start by listing outdated dependencies:

```shell
$ go list -u -m all
```

Be sure to read all changelogs of a dependency before updating. Then update it:

```shell
$ go get -u package/path@latest
# or
$ go get -u ./...
```

Then, tidy up the `go.mod`:

```shell
$ go mod tidy
```

Once your done, run all tests to ensure nothing broke and fix it otherwise.

## Update IP database

To update embedded IP database, simply run `scripts/update-geodb.sh` script from
repository's root and commit changes.

## Update User-Agent parser

To update user-agent parser regexes, run `scripts/update-uap.sh` script from
repository's root and commit changes.

> **NOTE**: Update script downloads latest regex file and patches it. If an error
occurred during the patch phase, fix it.

## Update User-Agent regexes patch

To update user-agent patch, start with a clean git working tree by stashing,
committing or discarding all your changes. Then, reverse apply the patch as follow:

```sh
$ git apply -R --reject --whitespace=fix pkg/embedded/uap/regexes.patch
```

Add these changes without committing them and then reapply the patch:

```sh
$ git add .
$ git apply pkg/embedded/uap/regexes.patch
```

Edit the regexes file and generate a new patch:

```sh
$ git diff pkg/embedded/uap/regexes.yml > pkg/embedded/uap/regexes.patch
```

That's it, you can commit your changes.

## Release a new version

Before releasing a new version, be sure to update dependencies, IP database
and maybe Grafana and ClickHouse Docker images to ensure compatibility with
latest versions.

Run benchmarks `make -C tests perf` to ensure there is no performance regression
in ClickHouse and Prisme.

Check `docker compose` based deployment in `deploy/` directory using
`prismelabs/analytics:dev` image.

Once everything is works as expected, update version in
`deploy/docker-compose.yml` and `flake.nix`.

