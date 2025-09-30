.PHONY: all
all: deno # perf must be called manually

.PHONY: deno
deno:
	$(MAKE) -C deno

deno/%: FORCE
	$(MAKE) -C deno $*

.PHONY: perf
perf:
	$(MAKE) -C perf

perf/%: FORCE
	$(MAKE) -C perf $*

.PHONY: lint
lint: deno/lint perf/lint

.PHONY: lint/fix
lint/fix: deno/lint/fix perf/lint/fix

FORCE:
