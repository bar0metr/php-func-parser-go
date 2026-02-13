BINARY := bin/phpfuncparser

.PHONY: all test build

all: test build

test:
	go test ./...

build:
	go build -o $(BINARY) ./cmd/phpfuncparser

