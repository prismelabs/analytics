.PHONY: start
start:
	go run ./cmd/server |& bunyan

watch/%:
	# When a new file is added, you must rerun make watch/...
	find . | entr -n -r sh -c "$(MAKE) $*"

.PHONY: lint
lint:
	golangci-lint run --timeout 2m ./...

.PHONY: test/unit
test/unit:
	go test -v ./...

.PHONY: test/e2e
test/e2e:
	@true

.PHONY: build
build:
	@true

.PHONY: docker/build
docker/build:
	@true
