repository_root := $(shell git rev-parse --show-toplevel)
repository_root := $(or $(repository_root), $(CURDIR))
include $(repository_root)/variables.mk

GENENV_FILES ?= $(wildcard ./config/*)
GENENV_FILE ?= ./config/genenv.local.sh

COMPOSE_PROJECT_NAME ?= $(notdir $(CURDIR))

.PHONY: start
start: start/server

start/%: .env
	$(MAKE) go/build/$*
	source ./.env \
	&& $(DOCKER_COMPOSE) \
		-f ./docker-compose.$${PRISME_MODE}.yml \
		up --wait
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.dev.yml \
		down
	-source ./.env \
	&& $(DOCKER_COMPOSE) \
		-f ./docker-compose.dev.yml \
		up --wait --force-recreate
	$(DOCKER) logs -f $(COMPOSE_PROJECT_NAME)-prisme-1 |& bunyan

.PHONY: stop
stop:
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.dev.yml \
		-f ./docker-compose.default.yml \
		-f ./docker-compose.ingestion.yml \
		stop

.PHONY: down
down:
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.dev.yml \
		-f ./docker-compose.default.yml \
		-f ./docker-compose.ingestion.yml \
		down

.PHONY: clean
clean:
	@touch .env
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.dev.yml \
		-f ./docker-compose.default.yml \
		-f ./docker-compose.ingestion.yml \
		 down --volumes --remove-orphans
	rm -f .env

watch/%:
	# When a new file is added, you must rerun make watch/...
	find '(' -regex '.*\.go$$' -or -regex '.*docker-compose.*' ')' \
		-and -not -regex '.*_test.go' \
		-and -not -regex '.*_gen.go' \
		-and -not -regex '.*/tests/.*' | \
		entr -n -r sh -c "$(MAKE) $*"

.PHONY: lint
lint: codegen
	golangci-lint run --timeout 2m ./...
	$(MAKE) -C ./tests lint

.PHONY: lint/fix
lint/fix:
	$(MAKE) -C ./tests lint/fix

.PHONY: codegen
codegen: ./pkg/embedded/static/wa.js
	wire ./...
	go generate -skip="wire" ./...

./pkg/embedded/static/wa.js: ./tracker/web_analytics.js
	minify --js-version 2019 $^ > $@

$(GENENV_FILE):
	@echo "$(GENENV_FILE) doesn't exist, generating one..."
	@printf '#!/usr/bin/env bash\n\nDIR="$$(dirname $$0)"\nsource "$$DIR/genenv.sh"\n\n# setenv PRISME_XXX_OPTION "value"' > $@
	@chmod +x $(GENENV_FILE)
	@echo "$(GENENV_FILE) generated, you can edit it!"

.env: $(GENENV_FILES) $(GENENV_FILE)
	bash $(GENENV_FILE) > .env; \

.PHONY: test/unit
test/unit: codegen
	go test -v -tags assert -short -race -bench=./... -benchmem ./...

.PHONY: test/integ
test/integ: .env
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.default.yml \
		up --wait
	source ./.env && go test -race -v -p 1 -run TestInteg ./...
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.default.yml \
		down --volumes --remove-orphans

.PHONY: test/e2e
test/e2e:
	$(MAKE) -C ./tests

tests/%: FORCE
	$(MAKE) -C ./tests $*

.PHONY: go/build
go/build: go/build/server

go/build/%: FORCE codegen
	go build -o prisme -race ./cmd/$*

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
