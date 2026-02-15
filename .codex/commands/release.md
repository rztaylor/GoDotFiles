---
description: Cut a new GDF release
---

1. Prepare changelog:
- Move `[Unreleased]` items in `CHANGELOG.md` to a new version section (for example `## [0.6.0] - 2026-02-15`).
- Confirm entries are user-facing (behavior/outcomes) and remove internal implementation details.

2. Commit release prep:
- Commit the changelog/version updates.

3. Tag and push release:
`make release VERSION=vX.Y.Z`

4. Verify GitHub Actions:
- Confirm `.github/workflows/release.yml` succeeds.

5. Optional local preview of release notes:
`go run ./scripts/extract-release-notes > /tmp/release_notes.md`

6. Post-release verification:
- Confirm GitHub release notes are correct
- Run `gdf update --check` from an installed binary
