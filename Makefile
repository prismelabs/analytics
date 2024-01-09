DOCKER ?= docker
DOCKER_COMPOSE ?= docker compose

GENENV_FILE ?= ./config/genenv.local.sh

.PHONY: start
start: dev/start
	source ./.env && go run ./cmd/server |& bunyan

.PHONY: dev/start
dev/start: .env
	source ./.env && $(DOCKER_COMPOSE) up -d

.PHONY: dev/stop
dev/stop:
	$(DOCKER_COMPOSE) stop

.PHONY: dev/down
dev/down:
	$(DOCKER_COMPOSE) down

.PHONY: dev/clean
dev/clean:
	$(DOCKER_COMPOSE) down --volumes --remove-orphans

$(GENENV_FILE):
	@echo "$(GENENV_FILE) doesn't exist, generating one..."
	@printf '#!/usr/bin/env bash\n\nDIR="$$(dirname $$0)"\nsource "$$DIR/genenv.sh"\n\n# setenv PRISME_XXX_OPTION "value"' > $@
	@chmod +x $(GENENV_FILE)
	@echo "$(GENENV_FILE) generated, you can edit it!"

.env: $(GENENV_FILE)
	bash $(GENENV_FILE) > .env; \

watch/%:
	# When a new file is added, you must rerun make watch/...
	find . | entr -n -r sh -c "$(MAKE) $*"

.PHONY: lint
lint:
	golangci-lint run --timeout 2m ./...

.PHONY: test/unit
test/unit: dev/start
	source ./.env && go test -v ./...

.PHONY: test/e2e
test/e2e:
	@true

.PHONY: build
build:
	@true

.PHONY: docker/build
docker/build:
	@true
