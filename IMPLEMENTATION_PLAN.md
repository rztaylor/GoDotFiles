# GDF (Go Dotfiles) - Implementation Plan

## Overview

**gdf** is a cross-platform dotfile manager written in Go that unifies **packages, configuration files, and shell aliases** into coherent "app bundles."

> For detailed documentation, see the [`docs/`](docs/) directory.

---

## Core Concept: App Bundles

Packages, dotfiles, and aliases belong together. When you install `kubectl`, you also want `~/.kube/config` managed, aliases like `k`, and shell completions loaded.

```
┌─────────────────────────────────────────────────────────┐
│                    App Bundle: kubectl                  │
├─────────────────────────────────────────────────────────┤
│   Package:      kubectl (brew/apt/dnf)                  │
│   Dotfiles:     ~/.kube/config                          │
│   Aliases:      k, kgp, kns, kaf, kl                    │
│   Companions:   kubectx, kubens, stern, k9s             │
└─────────────────────────────────────────────────────────┘
```

**Details**: See [Architecture Overview](docs/architecture/overview.md)

---

## Design Principles

### 80/20 CLI Philosophy

| Use Case | Approach |
|----------|----------|
| Add app, track config, add alias | **CLI** - one-liners |
| Custom hooks, complex conditions | **YAML** - full control |

**Details**: See [CLI Reference](docs/reference/cli.md)

---

## Architecture

```
CLI (cobra) → App Engine → Platform Layer → Git Storage
```

| Package | Responsibility |
|---------|----------------|
| `internal/cli` | Cobra commands |
| `internal/config` | YAML loading/validation |
| `internal/engine` | Core orchestration |
| `internal/apps` | App bundle operations |
| `internal/packages` | Package manager abstraction |
| `internal/shell` | Shell script generation |
| `internal/platform` | OS detection |
| `internal/git` | Git operations |

**Details**: See [Component Documentation](docs/architecture/components.md)

---

## Implementation Phases

### Phase 0: Project Setup ✅
- Go module structure
- Documentation templates
- GEMINI.md AI guidelines
- Gemini workflows

### Phase 1: Core (MVP) - Current
See [`TASKS.md`](TASKS.md) for detailed task breakdown.

1. App bundle structure & YAML format
2. Core commands: `init`, `add`, `track`, `apply`
3. Dotfile symlinking with backup
4. Package management (brew + apt)
5. Alias management
6. Profile management
7. Git integration
8. Tests (TDD)

### Phase 2: User Experience
- **UI Framework (Charmbracelet/Bubble Tea)**
- Core App Library (Embedded Recipes)
- Interactive wizards
- `gdf status` with rich output
- `gdf doctor` and `gdf fix`

### Phase 3: Advanced
- Remote Recipe Ecosystem
- Profile conditions (OS, hostname)
- Companion apps & plugins
- Template rendering
- Pre/post hooks

### Phase 4: Polish & AI
- Conflict resolution UI
- Secret management
- Fish shell support
- AI recipe generation

---

## Confirmed Decisions

| Decision | Choice |
|----------|--------|
| Command name | `gdf` |
| Storage location | `~/.gdf/` |
| Model | App-centric (package+dotfiles+aliases) |

---

## Documentation

| Document                                                | Purpose                |
| ------------------------------------------------------- | ---------------------- |
| [`docs/user-guide/`](docs/user-guide/)                  | How to use gdf         |
| [`docs/reference/`](docs/reference/)                    | CLI & config reference |
| [`ui-style-guide.md`](docs/reference/ui-style-guide.md) | UI & UX guidelines     |
| [`docs/architecture/`](docs/architecture/)              | Technical design       |
| [`TASKS.md`](TASKS.md)                                  | Current task list      |
| [`CHANGELOG.md`](CHANGELOG.md)                          | Version history        |
| [`GEMINI.md`](GEMINI.md)                                | AI development rules   |
