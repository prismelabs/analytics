.PHONY: all
all: all/events # all/trusted-proxy all/trusted-proxy-custom-header all/untrusted-proxy
	@true

all/%: FORCE start/% .WAIT test/% .WAIT clean/%
	@true

start/%: FORCE
	$(MAKE) -C $* start

test/%: FORCE
	$(MAKE) -C $* test

clean/%: FORCE
	$(MAKE) -C $* clean

.PHONY: lint
lint:
	bunx ts-standard

.PHONY: lint/fix
lint/fix:
	bunx ts-standard --fix

FORCE:
