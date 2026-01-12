BINARY=bin/phpfuncparser

.PHONY: test build

test:
	go mod tidy
	go test ./...

build:
	go mod tidy
	go build -o $(BINARY) ./cmd/phpfuncparser
