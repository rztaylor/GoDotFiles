# Codex Workspace Guide

This folder defines fast-path operating commands for working in GDF with minimal prompt churn.

## Project Objective

GDF is a Go-based, cross-platform dotfile manager centered on App Bundles (packages + dotfiles + aliases), with profile composition and Git-backed sync.

## Automation Mode

- Default behavior: execute requested changes end-to-end without asking for per-file or per-edit confirmation.
- Interrupt only for destructive actions not explicitly requested.
- Interrupt for ambiguous product decisions that materially change behavior.
- Interrupt for sandbox/permission escalation requests.
- If worktree is dirty, avoid unrelated files and continue within request scope.

## Default Working Assumptions

- Respect `AGENTS.md` as source of truth for engineering process.
- Prefer existing abstractions over new implementations (`internal/util`, package managers, resolver/linker patterns).
- For behavior changes: update tests first, then implementation, then docs/changelog/tasks as required.
- Avoid touching unrelated files in a dirty worktree.

## Fast Code Discovery

- List packages quickly:
`rg --files internal | sed 's|/[^/]*$||' | sort -u`
- Find CLI command definitions:
`rg "^func new.*Cmd|Use:\\s+\"" internal/cli`
- Find schema/versioned YAML usage:
`rg "kind:\\s+[A-Za-z]+/v[0-9]+" internal docs`
- Find engine orchestration entry points:
`rg "func \\(.*\\) (Apply|Restore|Link|Resolve)" internal/engine internal/apps`
- Find tests near target code:
`rg --files internal | rg "_test\\.go$"`

## Command Shortcuts

- Build: `.codex/commands/build.md`
- Context: `.codex/commands/context.md`
- Review: `.codex/commands/review.md`
- Test: `.codex/commands/test.md`
- Lint: `.codex/commands/lint.md`
- Docs sync: `.codex/commands/docs-update.md`
- Release: `.codex/commands/release.md`

## Prompt Templates

- Feature implementation:
`Run .codex/commands/context.md, implement <feature>, add/update tests, update docs/changelog/tasks if needed, then run go test ./...`
- Bug fix:
`Run .codex/commands/context.md, reproduce and fix <bug>, add regression test, run go test ./...`
- Code review:
`Run .codex/commands/review.md and provide findings ordered by severity with file references.`
- Docs sync:
`Run .codex/commands/docs-update.md and patch any drift you find.`
- Release prep:
`Run .codex/commands/test.md, .codex/commands/lint.md, and .codex/commands/release.md; report blockers.`

## High-Signal Docs

- `README.md`
- `docs/architecture/overview.md`
- `docs/architecture/components.md`
- `docs/architecture/versioning.md`
- `docs/reference/cli.md`
- `docs/reference/yaml-schemas.md`
- `TASKS.md`
- `CHANGELOG.md`
