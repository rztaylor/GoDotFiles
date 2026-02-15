---
description: Run all tests and coverage checks
---

1. Run the repo-standard test runner:
`make test`

2. Run baseline package tests expected by `AGENTS.md`:
`go test ./...`

3. Optional deeper checks:
- `go test -race ./...`
- `go test -coverprofile=coverage.out ./...`
- `go tool cover -func=coverage.out`
