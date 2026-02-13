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

5. Check for speculative comments:
   - Ensure comments are factual and professional (no "I think", "maybe", "TODO: fix later" without ticket)
   - Ensure comments explain WHY, not HOW
   - Remove any "chain-of-thought" or debug comments
