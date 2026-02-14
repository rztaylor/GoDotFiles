# GDF - AI Development Guidelines

## Project Context

GDF (Go Dotfiles) is a cross-platform dotfile manager written in Go. It uses the **App Bundle** concept to unify packages, dotfiles, and shell aliases into coherent units.

## Development Principles

### 1. Test-Driven Development (TDD)

- Every PR must include tests for new functionality
- Tests must pass before any code is merged
- Use table-driven tests for Go code
- Target 80%+ coverage for core packages

```go
// Example: table-driven test pattern
func TestDetectOS(t *testing.T) {
    tests := []struct {
        name     string
        procFile string
        want     string
    }{
        {"Linux", "", "linux"},
        {"WSL", "Microsoft", "wsl"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### 2. Documentation Updates

When modifying code, **ALWAYS** update relevant documentation:

| Change Type | Update Required |
|-------------|-----------------|
| New CLI command | `docs/reference/cli.md`, relevant user guide |
| New config option | `docs/reference/config.md`, `docs/reference/yaml-schemas.md` |
| Architecture change | `docs/architecture/components.md`, `docs/architecture/decisions.md` |
| New internal package | `docs/architecture/components.md`, add `doc.go` |
| Bug fix | Consider FAQ update if user-facing |

### 3. Component Documentation

Each internal package **must** have a `doc.go` file containing:

- Package purpose and responsibility
- Key types and their roles
- Usage examples
- Dependencies on other packages

### 4. Avoiding Duplication

Before implementing new functionality:

1. Search existing code for similar patterns
2. Check `docs/architecture/components.md` for existing capabilities
3. Reuse existing utilities from `internal/util/`
4. If similar code exists, refactor to share

### 5. Code Quality Checklist

Before completing any change:

- [ ] All tests pass (`go test ./...`)
- [ ] Linting passes (`golangci-lint run`)
- [ ] Documentation updated
- [ ] UI follows [Style Guide](docs/reference/ui-style-guide.md) (if applicable)
- [ ] No functionality duplication
- [ ] Error messages are user-friendly
- [ ] CLI help text is clear and complete

### 6. Professional Code Comments

- Comments must be **factual** and **professional**.
- Explain **WHY** code exists, not **HOW** it works (the code explains the how).
- **DO NOT** write "thinking" or "chain-of-thought" comments (e.g., `// I think this should work`, `// trying to fix bug`).
- **DO NOT** use speculative language (e.g., "maybe", "probably", "unsure").
- Remove commented-out code before merging.

### 7. Schema Versioning

All YAML files (except simple lists like `tasks.md`) MUST include a `kind` field following the format `<Type>/<Version>` (e.g., `App/v1`).
Refer to `docs/architecture/versioning.md` for details.

## File Organization

### Key Directories

| Directory | Responsibility |
|-----------|----------------|
| `cmd/gdf/` | CLI entry point |
| `internal/cli/` | Cobra command implementations |
| `internal/config/` | Configuration loading and validation |
| `internal/engine/` | Core business logic, orchestration |
| `internal/apps/` | App bundle operations |
| `internal/packages/` | Package manager integrations (brew, apt, dnf) |
| `internal/shell/` | Shell script generation |
| `internal/platform/` | OS detection and path abstraction |
| `internal/git/` | Git repository operations |
| `internal/util/` | Shared utilities |

### Adding a New Feature

1. Create/update tests in `*_test.go`
2. Implement in appropriate `internal/` package
3. Ensure data structures implement `schema.TypeMeta`
4. Add CLI command in `internal/cli/`
5. Update docs in `docs/`
6. Run full test suite

## Library Recipes

Recipes in `internal/library/` should:

- Include package definitions for brew, apt, dnf minimum
- Include `kind: Recipe/v1` in the YAML definition
- Suggest common aliases (keep it reasonable, not exhaustive)
- Include shell completions if available
- Be tested on macOS and Ubuntu

## Commit Messages

Use conventional commits:

```
feat(cli): add gdf track command
fix(packages): handle brew cask packages correctly
docs(readme): update installation instructions
test(apps): add tests for bundle loading
```

---

## Project File Maintenance

### Critical Files to Keep Updated

| File | Purpose | Update Trigger |
|------|---------|----------------|
| `TASKS.md` | Pending/in-progress tasks | When starting or completing tasks |
| `CHANGELOG.md` | History of changes | When completing features or fixes |
| `IMPLEMENTATION_PLAN.md` | Design decisions | When architecture changes |
| `release` | Automated release | When ready to ship |

### Accidental Release Prevention

The `make release` command includes checks to prevent accidental releases (e.g., ensuring `VERSION` is set, git is clean). Never push tags manually unless you know what you are doing.

### Release Workflow

1. **Prepare**: Update `CHANGELOG.md` (unreleased -> version)
2. **Tag**: `make release VERSION=vX.Y.Z`
3. **Build**: Watch GitHub Actions


### TASKS.md Rules

1. **Only pending and in-progress tasks** - No completed tasks
2. **Mark in-progress**: Change `[ ]` to `[/]` when starting work
3. **Remove when done**: Delete completed tasks from this file
4. **Pick top to bottom**: Work on tasks in order within each section

### CHANGELOG.md Rules

1. **Add entries under `[Unreleased]`** when completing work
2. **Use correct category**: Added, Changed, Fixed, etc.
3. **Imperative mood**: "Add feature" not "Added feature"
4. **Move to version section** on release

### Workflow for Completing a Task

```
1. Mark task as [/] in TASKS.md
2. Implement with tests (TDD)
3. Update relevant docs
4. Remove task from TASKS.md
5. Add entry to CHANGELOG.md under [Unreleased]
6. Commit with conventional commit message
```

### IMPLEMENTATION_PLAN.md Rules

1. **Keep architecture current** - Update when design changes
2. **Mark phase progress** - Check off completed phases
3. **Add new decisions** - Document significant choices

