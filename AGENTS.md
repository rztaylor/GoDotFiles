# GDF - Codex Agent Guide

## Project Context

GDF (Go Dotfiles) is a cross-platform dotfile manager written in Go. It uses the App Bundle concept to combine packages, dotfiles, and aliases.

## Engineering Rules

1. Test-driven changes
- Add or update tests for every behavior change.
- Prefer table-driven tests in Go.
- Run `go test ./...` before finishing.

2. Keep docs in sync
- New CLI command: update `docs/reference/cli.md` and relevant user docs.
- New config option: update `docs/reference/yaml-schemas.md`.
- Architecture changes: update `docs/architecture/components.md` and `docs/architecture/decisions.md`.
- New internal package: add/update `doc.go` and `docs/architecture/components.md`.

3. Package documentation
- Every `internal/*` package must keep a `doc.go` describing purpose, key types, and dependencies.

4. Avoid duplication
- Search for existing implementations before adding new code.
- Reuse utilities from `internal/util/` where possible.

5. Code quality checklist
- Tests pass.
- Linting passes (`golangci-lint run`).
- Formatting clean (`gofmt`).
- User-facing errors and help text are clear.
- No duplicated functionality.

6. Comment quality
- Comments should explain why, not how.
- Keep comments factual and professional.
- Do not leave speculative/debug comments or commented-out code.

7. YAML schema versioning
- YAML definitions (except simple lists) must include `kind: <Type>/<Version>` such as `Recipe/v1`.
- Follow `docs/architecture/versioning.md`.

## File Responsibilities

- `cmd/gdf/`: CLI entrypoint
- `internal/cli/`: cobra commands
- `internal/config/`: config parsing and validation
- `internal/engine/`: orchestration/business logic
- `internal/apps/`: app bundle behavior
- `internal/packages/`: package manager integrations
- `internal/shell/`: shell script generation
- `internal/platform/`: platform/path logic
- `internal/git/`: git operations
- `internal/util/`: shared helpers

## Project Maintenance

- `TASKS.md`: only pending/in-progress items.
- `CHANGELOG.md`: keep `[Unreleased]` updated for user-facing changes.
- `IMPLEMENTATION_PLAN.md`: update with design/phase changes.

## Task Completion Flow

1. Mark task in-progress in `TASKS.md` (`[ ]` -> `[/]`).
2. Implement with tests.
3. Update docs.
4. Remove completed task from `TASKS.md`.
5. Add changelog entry under `[Unreleased]`.
6. Use a conventional commit message.
