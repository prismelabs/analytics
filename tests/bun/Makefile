.PHONY: all
all: all/default all/events all/trusted-proxy all/untrusted-proxy
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
