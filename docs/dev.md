# Developer documentation

## Local development

In order to run Prisme locally, you must have the
[Nix](https://nixos.org/download/) package manager installed.

Then, you can enter development shell as follow:

```shell
$ nix develop
```

This will start a new shell with all required tools installed (in an isolated
directory named `/nix/store` by default).

### Build docker image

Before starting development environment, you must build Prisme docker image at
least once.

```shell
$ make docker/build
```

Prisme uses [Nix](https://nixos.org) to build ultra light docker image that are
**reproducible**. Building twice the image should produce the **exact** same
images.

Nix builds are slower than `docker build` but don't worry you don't have to
rebuild the entire image at each changes?

### Start development environment

You can now start development environment.

```shell
$ make start
# or
$ make watch/start
```

Target `watch/start` will watches all existing Go and docker compose files
and restart the service on changes.

> **NOTE**: Files created after `watch/start` don't trigger restarts.

Be sure to read the `Makefile` to read and understand all targets.

## Documentation

All packages are documented. You can start a local documentation server using
`godoc -http=:6060`.

## Testing

Prisme is designed to be a robust product so we works hard to have good tests
that cover critical parts of the service. This includes but isn't limited to:
* Core features
* Security related features
* Observability (metrics, logs)

Depending on the feasibility, some tests are written as unit tests, integration
tests or end-to-end tests.

That being said, we favor end-to-end tests when possible as they're closest
to how Prisme is used in production.

### Unit tests

Unit tests follows Go convention and are placed in `*_test.go` files next to
other go files and uses `github.com/stretchr/testify` for asserts.

You can run unit tests as follows:

```shell
$ make test/unit
```

### Integration tests

Integration tests are also placed in `*_test.go` files along unit tests but
follows a specific convention:

```go
// Integration tests function starts with TestInteg.
func TestIntegXXX(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Your test here.
```

You can run integration tests as follow.

```shell
$ make test/integ
```

### End-to-end tests

Finally, end-to-end tests are stored under `tests/`.

```shell
$ make test/e2e
```

