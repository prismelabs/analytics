.PHONY: all
all: setup .WAIT all/events all/trusted-proxy all/trusted-proxy-custom-header all/untrusted-proxy
	@true

all/%: start/% .WAIT test/% .WAIT clean/%
	@true

start/%:
	$(MAKE) -C $* start

test/%:
	$(MAKE) -C $* test

clean/%:
	$(MAKE) -C $* clean

.PHONY: lint
lint:
	bunx ts-standard

.PHONY: lint/fix
lint/fix:
	bunx ts-standard --fix

setup: node_modules

node_modules:
	$(DOCKER) run --rm -v $$PWD:/app -w /app docker.io/oven/bun bun install

