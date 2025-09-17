DIST_DIR ?= dist

build: $(DIST_DIR)/index.html

$(DIST_DIR)/index.html: $(shell find src -type f) vite.config.ts index.html deno.json deno.lock
	deno task build --outDir $(DIST_DIR)
