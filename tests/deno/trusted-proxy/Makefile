repository_root := $(shell git rev-parse --show-toplevel)
export repository_root := $(or $(repository_root), $(CURDIR))
include $(repository_root)/variables.mk

export COMPOSE_PROJECT_NAME := prisme-deno-trusted-proxy

.PHONY: start
start:
	source $(CURDIR)/prisme.env && \
	$(DOCKER_COMPOSE) \
		-f ../docker-compose.yml \
		-f ./docker-compose.env.yml \
		up --wait -d

.PHONY: test
test:
	$(DOCKER_COMPOSE) \
		-f ./docker-compose.test.yml \
		up \
		--abort-on-container-exit \
		--exit-code-from deno

.PHONY: clean
clean:
	$(DOCKER_COMPOSE) \
		-f ../docker-compose.yml \
		-f ./docker-compose.env.yml \
		-f ./docker-compose.test.yml \
		down --volumes --remove-orphans
