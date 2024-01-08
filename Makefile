.PHONY: start
start:
	go run ./cmd/server

.PHONY: lint
lint:
	golangci-lint run --timeout 2m ./...

.PHONY: test/unit
test/unit:
	go test -v -p 1 -count=1 ./...

.PHONY: test/e2e
test/e2e:
	@true

.PHONY: build
build:
	@true

.PHONY: docker/build
docker/build:
	@true
