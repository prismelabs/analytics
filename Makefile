repository_root := $(shell git rev-parse --show-toplevel)
repository_root := $(or $(repository_root), $(CURDIR))
include $(repository_root)/variables.mk

GENENV_FILES ?= $(wildcard ./config/*)
GENENV_FILE ?= ./config/genenv.local.sh

COMPOSE_PROJECT_NAME ?= $(notdir $(CURDIR))

GO ?= go

default: start

.PHONY: start
start: tmp/.env tmp/prisme
	source ./tmp/.env \
	&& $(DOCKER_COMPOSE) \
		-f ./docker-compose.yml \
		up --wait
	source ./tmp/.env \
		&& air --build.cmd '$(MAKE) tmp/prisme' --build.bin './tmp/prisme' \
		|& bunyan

.PHONY: stop
stop:
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.yml \
		stop

.PHONY: down
down:
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.yml \
		down

.PHONY: clean
clean:
	-$(DOCKER_COMPOSE) \
		-f ./docker-compose.yml \
		 down --volumes --remove-orphans
	rm -rf tmp/

.PHONY: lint
lint: codegen
	golangci-lint run --allow-parallel-runners --timeout 2m ./...
	$(MAKE) -C ./tests lint

.PHONY: lint/fix
lint/fix:
	$(MAKE) -C ./tests lint/fix

.PHONY: codegen
codegen: ./pkg/embedded/static/wa.js

./pkg/embedded/static/wa.js: ./tracker/web_analytics.js
	minify --js-version 2019 $^ > $@

$(GENENV_FILE):
	@echo "$(GENENV_FILE) doesn't exist, generating one..."
	@printf '#!/usr/bin/env bash\n\nDIR="$$(dirname $$0)"\nsource "$$DIR/genenv.sh"\n\n# setenv PRISME_XXX_OPTION "value"' > $@
	@chmod +x $(GENENV_FILE)
	@echo "$(GENENV_FILE) generated, you can edit it!"

tmp/.env: tmp/ $(GENENV_FILES) $(GENENV_FILE)
	bash $(GENENV_FILE) > tmp/.env; \

tmp/prisme: go/build/prisme

tmp/:
	mkdir -p tmp/

.PHONY: test/unit
test/unit: codegen
	go test -v -tags assert -short -race -bench=./... -benchmem ./...

.PHONY: test/integ
test/integ: tmp/.env
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.yml \
		up --wait
	source ./tmp/.env && go test -tags chdb -v -race -p 1 -run TestInteg ./...
	source ./tmp/.env && go test -tags chdb -v -p 1 -run TestIntegNoRaceDetector ./...
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.yml \
		down --volumes --remove-orphans

.PHONY: test/e2e
test/e2e:
	$(MAKE) -C ./tests

tests/%: FORCE
	$(MAKE) -C ./tests $*

.PHONY: go/build
go/build: go/build/prisme

go/build/%: FORCE codegen
	$(GO) build -o ./tmp/$* -race ./cmd/$*

.PHONY: nix/build
nix/build:
	nix build -L .#default

.PHONY: docker/build
docker/build: docker/build/prisme
	@true

.PHONY: docker/build/prisme
docker/build/prisme:
	nix build -L .#docker
	$(DOCKER) load < result
	if [ "$${REMOVE_RESULT:=1}" = "1" ]; then rm -f result; fi

.PHONY: docker/build/clickhouse
docker/build/clickhouse:
	$(DOCKER) build $(repository_root)/docker/clickhouse -t prismelabs/clickhouse
	tag=$$(grep 'FROM' docker/clickhouse/Dockerfile | sed -E 's/.*[a-zA-Z]+:(.*)/\1/'); \
		docker tag prismelabs/clickhouse prismelabs/clickhouse:$$tag

FORCE:
