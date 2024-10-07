# Prisme server

This directory holds `main` package of Prisme server.

## Dependency Injection

Dependencies are injected using [`wire`](https://github.com/google/wire). There,
is a `wire.go` file per Prisme mode (`default` or `ingestion`) in their
respective packages.

