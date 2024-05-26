repository_root := $(shell git rev-parse --show-toplevel)
export repository_root := $(or $(repository_root), $(CURDIR))
include $(repository_root)/variables.mk

.PHONY: all
all: ingestion query

.PHONY: ingestion
ingestion:
	$(MAKE) -C ingestion

ingestion/%: FORCE
	$(MAKE) -C ingestion $*

.PHONY: query
query:
	$(MAKE) -C query

query/%: FORCE
	$(MAKE) -C query $*

FORCE:
