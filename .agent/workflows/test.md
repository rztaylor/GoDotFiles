---
description: Run all tests and check coverage
---
// turbo-all
1. Run the full test suite:
   `go test -v -cover ./...`

2. Check for race conditions:
   `go test -race ./...`

3. Generate and open coverage report:
   `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`
