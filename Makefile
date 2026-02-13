.PHONY: all build test lint clean coverage install

BINARY_NAME=gdf

all: build test

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X github.com/rztaylor/GoDotFiles/internal/cli.Version=$(VERSION) \
           -X github.com/rztaylor/GoDotFiles/internal/cli.Commit=$(COMMIT) \
           -X github.com/rztaylor/GoDotFiles/internal/cli.Date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/gdf

install: build
	mkdir -p $(HOME)/bin
	cp $(BINARY_NAME) $(HOME)/bin/

test:
	go run scripts/test_runner.go

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out
