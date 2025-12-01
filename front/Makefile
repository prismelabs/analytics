DIST_DIR ?= dist

.PHONY: start
start:
	deno task dev

.PHONY: build
build: $(DIST_DIR)/index.html

$(DIST_DIR)/index.html: $(shell find src -type f) vite.config.ts index.html deno.json deno.lock
	deno task build --outDir $(DIST_DIR)

.PHONY: lint
lint:
	deno fmt --check .

.PHONY: lint/fix
lint/fix:
	deno fmt .
