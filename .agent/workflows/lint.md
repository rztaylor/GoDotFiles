---
description: Run linting and formatting checks
---
// turbo-all
1. Run golangci-lint:
   `golangci-lint run`

2. Check formatting:
   `gofmt -d .`

3. Run go vet:
   `go vet ./...`

4. Check for tidied modules:
   `go mod tidy && git diff --exit-code go.mod go.sum`
