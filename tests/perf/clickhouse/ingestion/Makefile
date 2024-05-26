repository_root := $(shell git rev-parse --show-toplevel)
export repository_root := $(or $(repository_root), $(CURDIR))
include $(repository_root)/variables.mk

COMPOSE_PROJECT_NAME ?= prisme-perf
export COMPOSE_PROJECT_NAME

.PHONY: all
all: start test clean

.PHONY: start
start:
	$(DOCKER_COMPOSE) \
		-f ../../docker-compose.env.yml \
		up clickhouse --wait -d

.PHONY: test
test:
	$(MAKE) -C $(repository_root) go/build/addevents
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.yml \
		up --detach
	$(DOCKER) logs -f $(COMPOSE_PROJECT_NAME)-addevents-1 |& bunyan

.PHONY: clean
clean:
	$(DOCKER_COMPOSE) \
		-f ../../docker-compose.env.yml \
		down --remove-orphans --volumes
