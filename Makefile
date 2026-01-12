BINARY=bin/phpfuncparser

.PHONY: test build

test:
	go test ./...

build:
	go build -o $(BINARY) ./cmd/phpfuncparser
