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

install: test build
	mkdir -p $(HOME)/bin
	cp $(BINARY_NAME) $(HOME)/bin/

test:
	GOCACHE=$${GOCACHE:-/tmp/go-build} go run ./scripts/test-runner

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Release target
# Usage: make release VERSION=v0.6.0
release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is not set. Usage: make release VERSION=vX.Y.Z"; \
		exit 1; \
	fi
	@if ! echo "$(VERSION)" | grep -qE "^v[0-9]+\.[0-9]+\.[0-9]+"; then \
		echo "Error: VERSION must be in format vX.Y.Z"; \
		exit 1; \
	fi
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: git working directory is not clean. Commit or stash changes first."; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "Release $(VERSION) tagged and pushed. GitHub Action should trigger shortly."
