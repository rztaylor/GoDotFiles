---
description: Cut a new GDF release
---

1. Prepare changelog:
- Move `[Unreleased]` items in `CHANGELOG.md` to a new version section (for example `## [0.6.0] - 2026-02-15`).
- Confirm entries are user-facing (behavior/outcomes) and remove internal implementation details.

2. Run preflight checks:
- Confirm clean working tree: `git status --short`
- Confirm tag does not already exist locally: `git tag --list vX.Y.Z`
- Optional remote check: `git ls-remote --tags origin vX.Y.Z`

3. Commit release prep:
- Generate release notes to a temp file outside the repo:
`go run ./scripts/extract-release-notes > /tmp/release_notes.txt`
- Commit using that file:
`git add -A && git commit -F /tmp/release_notes.txt`

4. Tag and push release:
`make release VERSION=vX.Y.Z`

5. Recovery if tag exists locally after release failure:
- If `make release` created the local tag but push failed, push existing tag directly:
`git push origin vX.Y.Z`

6. Verify GitHub Actions:
- Confirm `.github/workflows/release.yml` succeeds.

7. Optional local preview of release notes:
`go run ./scripts/extract-release-notes > /tmp/release_notes.md`

8. Post-release verification:
- Confirm GitHub release notes are correct
- Run `gdf update --check` from an installed binary
