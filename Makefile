.PHONY: start
start:
	go run ./cmd/server

.PHONY: test
test:
	go test -v -p 1 -count=1 ./...

