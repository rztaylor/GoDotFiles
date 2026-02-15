---
description: Standardized code review checks
---

1. Scope the review from current changes:
`git status --short && git diff --stat`

2. Inspect full diff with file context:
`git diff`

3. Check for high-risk patterns:
`rg -n "TODO|FIXME|panic\\(|os\\.Exit\\(|fmt\\.Print|log\\.Print|http\\.Get\\(|exec\\.Command\\(" internal cmd`

4. Validate test impact:
`rg --files | rg "_test\\.go$" && go test ./...`

5. Validate lint and formatting:
- `make lint`
- `gofmt -l $(rg --files -g '*.go')`

6. Review docs/changelog/task hygiene for user-facing changes:
- `docs/reference/cli.md`
- `docs/reference/yaml-schemas.md`
- `docs/architecture/components.md`
- `CHANGELOG.md`
- `TASKS.md`

7. Produce findings by severity with file references:
- Critical: correctness/security/data loss
- High: behavior regressions or missing safeguards
- Medium: maintainability/test gaps
- Low: style/clarity issues
