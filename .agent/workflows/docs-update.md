---
description: Ensure documentation is in sync with code
---
1. Check that all CLI commands in `internal/cli/` are documented in `docs/reference/cli.md`

2. Verify config options match `docs/reference/config.md` and `docs/reference/yaml-schemas.md`
   - Ensure `kind` fields are documented for all schemas

3. Ensure `docs/architecture/components.md` reflects current package structure:
   - Compare with `ls internal/`
   - Each package should have a doc.go file

4. Update CHANGELOG.md if changes are user-facing

5. Check all internal links in docs are valid:
   `find docs -name "*.md" -exec grep -l "\[.*\](.*\.md)" {} \;`
