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
		-f ../docker-compose.env.yml \
		up --wait -d

.PHONY: test
test:
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.test.yml \
		up \
		--abort-on-container-exit \
		--exit-code-from k6

.PHONY: clean
clean:
	$(DOCKER_COMPOSE) \
		-f ../docker-compose.env.yml \
		-f ./docker-compose.test.yml \
		down --volumes --remove-orphans
