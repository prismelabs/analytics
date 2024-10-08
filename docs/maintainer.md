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
go mod tidy
```

Once your done, run all tests to ensure nothing broke and fix it otherwise.

## Update IP database

To generate a new IP database, clone
[`negrel/geoacumen-country`](https://github.com/negrel/geoacumen-country)
repository and run `make clean ip2asn-combined.mmdb`. Then copy
`ip2asn-combined.mmdb` to `pkg/embedded/geodb` and commit changes.

## Release a new version

Before releasing a new version, be sure to update dependencies, IP database
and maybe Grafana and ClickHouse Docker images to ensure compatibility with
latest versions.

