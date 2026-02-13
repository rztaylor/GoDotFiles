.PHONY: all build test lint clean coverage install

BINARY_NAME=gdf

all: build test

build:
	go build -o $(BINARY_NAME) ./cmd/gdf

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
