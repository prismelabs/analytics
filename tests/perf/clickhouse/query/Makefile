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
		-f ../../docker-compose.env.yml \
		up clickhouse --wait -d
	source ../../prisme.env \
	&& $(DOCKER) exec ${COMPOSE_PROJECT_NAME}-clickhouse-1 clickhouse-client \
		--format Pretty -q "SELECT \
			name, \
			formatReadableQuantity(total_rows) AS rows, \
			formatReadableSize(total_bytes) AS size, \
			formatReadableSize(total_bytes_uncompressed) AS uncompressed_size \
		FROM system.tables \
		WHERE database = '$${PRISME_CLICKHOUSE_DB}'"

.PHONY: test
test:
	set -e \
	&& source ../../prisme.env \
	&& for query in ./queries/*.sh; do\
		echo "$$query" \
		&& bash "$$query" \
		&& $(DOCKER) exec ${COMPOSE_PROJECT_NAME}-clickhouse-1 clickhouse-client \
			--format Pretty --database "$$PRISME_CLICKHOUSE_DB" \
			-q "$$(bash "$$query") LIMIT 10" \
		&& $(DOCKER) exec ${COMPOSE_PROJECT_NAME}-clickhouse-1 clickhouse-benchmark \
			--database "$$PRISME_CLICKHOUSE_DB" -i 10 -q "$$(bash "$$query")"; \
	done

.PHONY: clean
clean:
	$(DOCKER_COMPOSE) \
		-f ../../docker-compose.env.yml \
		down --remove-orphans --volumes
