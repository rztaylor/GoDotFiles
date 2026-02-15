---
description: Run lint and formatting checks
---

1. Run project lint target:
`make lint`

2. Ensure Go formatting is clean:
`gofmt -w $(rg --files -g '*.go')`

3. Re-run tests after lint/format changes:
`go test ./...`
