COMPOSE_PROJECT_NAME ?= prisme-perf
export COMPOSE_PROJECT_NAME

.PHONY: all
all: clickhouse/ingestion/start clickhouse/ingestion/test clickhouse/query/start clickhouse/query/test # clickhouse/ingestion/clean prisme/clean clickhouse/query/clean

.PHONY: clickhouse
clickhouse:
	$(MAKE) -C clickhouse

clickhouse/%: FORCE
	$(MAKE) -C clickhouse $*

.PHONY: prisme
prisme: FORCE
	$(MAKE) -C prisme

prisme/%: FORCE
	$(MAKE) -C prisme $*

.PHONY: lint
lint:
	deno fmt --check

.PHONY: lint/fix
lint/fix:
	deno fmt

FORCE:
