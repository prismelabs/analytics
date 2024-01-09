.PHONY: all
all: bun

.PHONY: bun
bun:
	$(MAKE) -C bun

bun/%:
	$(MAKE) -C bun $*

.PHONY: lint
lint: bun/lint

.PHONY: lint/fix
lint/fix: bun/lint/fix
