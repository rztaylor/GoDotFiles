---
description: Verify docs are in sync with code
---

1. Check CLI docs against `internal/cli/` and `docs/reference/cli.md`:
`rg "^func new.*Cmd|Use:\\s+\"" internal/cli`

2. Verify YAML schema docs match implementation:
- `docs/reference/yaml-schemas.md`
- `docs/architecture/versioning.md`
- `internal/schema/`
- `internal/config/`

3. Confirm architecture docs match package structure:
- Compare `internal/` to `docs/architecture/components.md`
- Ensure each internal package has `doc.go`

4. Confirm user docs are consistent with behavior:
- `README.md`
- `docs/user-guide/getting-started.md`
- `docs/user-guide/tutorial.md`

5. Update `CHANGELOG.md` for user-facing changes.

6. Check markdown links in docs:
`find docs -name "*.md" -exec grep -nE "\\[[^]]+\\]\\([^)]+\\.md\\)" {} \\;`
