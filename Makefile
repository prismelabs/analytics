repository_root := $(shell git rev-parse --show-toplevel)
repository_root := $(or $(repository_root), $(CURDIR))
include $(repository_root)/variables.mk

CONFIG_FILES ?= $(wildcard ./config/*.ini)
CONFIG_FILE ?= ./config/local.ini

COMPOSE_PROJECT_NAME ?= $(notdir $(CURDIR))

GO ?= go

default: start/prisme

.PHONY: start
start: tmp/.env
	source ./tmp/.env \
	&& $(DOCKER_COMPOSE) \
		-f ./docker-compose.yml \
		up --wait \

.PHONY: start/prisme
start/prisme: start tmp/prisme
	PRISME_CONFIG=$(CONFIG_FILE) air --build.cmd '$(MAKE) tmp/prisme' --build.bin './tmp/prisme' \
	|& bunyan

.PHONY: start/addevents
start/addevents: tmp/.env
	source ./tmp/.env \
	&& $(DOCKER_COMPOSE) \
		-f ./docker-compose.yml \
		up --wait \
	&& $(GO) run -tags test ./cmd/addevents $(ARGS) |& bunyan

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
	openapi-spec-validator openapi.yml
	golangci-lint run --allow-parallel-runners --timeout 2m ./...
	$(MAKE) -C ./tests lint

.PHONY: lint/fix
lint/fix:
	go fmt ./...
	$(MAKE) -C ./tests lint/fix

.PHONY: codegen
codegen: ./pkg/embedded/static/wa.js ./pkg/embedded/static/openapi.json ./pkg/embedded/dashboard

.PHONY: ./pkg/embedded/dashboard
./pkg/embedded/dashboard:
	mkdir -p $@
	$(MAKE) -C front DIST_DIR="../$@" build

./pkg/embedded/static/openapi.json: ./openapi.yml
	yq < $^ > $@

./pkg/embedded/static/wa.js: ./tracker/web_analytics.js
	minify --js-version 2019 $^ > $@

$(CONFIG_FILE): ./config/example.ini
	@echo "$(CONFIG_FILE) doesn't exist, generating one..."
	@cp ./config/example.ini $@
	@chmod +x $(CONFIG_FILE)
	@echo "$(CONFIG_FILE) generated, you can edit it!"

tmp/prisme: go/build/prisme

tmp/:
	mkdir -p tmp/

.PHONY: test/unit
test/unit: codegen
	go test -v -tags assert,test -short -race -bench=./... -benchmem ./...

.PHONY: test/integ
test/integ: start $(CONFIG_FILE)
	source ./tmp/.env && go test -tags chdb,test -v -race -p 1 -run TestInteg ./...
	source ./tmp/.env && go test -tags chdb,test -v -p 1 -run TestIntegNoRaceDetector ./...
	$(MAKE) clean

.PHONY: test/e2e
test/e2e:
	$(MAKE) -C ./tests

tests/%: FORCE
	$(MAKE) -C ./tests $*

.PHONY: phony/integ
bench/integ:
	source ./tmp/.env && go test -tags chdb,test -v -p 1 -run ^$$ -bench=BenchmarkInteg -benchtime 10s ./...

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

FORCE:
