.PHONY: all
all: bun # perf must be called manually

.PHONY: bun
bun:
	$(MAKE) -C bun

bun/%: FORCE
	$(MAKE) -C bun $*

.PHONY: perf
perf:
	$(MAKE) -C perf

perf/%: FORCE
	$(MAKE) -C perf $*

.PHONY: lint
lint: bun/lint perf/lint

.PHONY: lint/fix
lint/fix: bun/lint/fix perf/lint/fix

FORCE:
