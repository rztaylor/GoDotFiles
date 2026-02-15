# Component Responsibilities

Each internal package has a specific responsibility.

## Package Reference

| Package | Responsibility |
|---------|----------------|
| `internal/cli` | Cobra commands, input parsing, output formatting |
| `internal/config` | Configuration and YAML loading/validation |
| `internal/engine` | Core orchestration and business logic |
| `internal/apps` | App bundle CRUD, companion/plugin management |
| `internal/packages` | Package manager abstractions (brew, apt, dnf) |
| `internal/shell` | Shell script generation, aliases, functions, startup init tasks |
| `internal/platform` | OS detection, path normalization |
| `internal/git` | Git operations (clone, commit, push, pull) |
| `internal/state` | Applied profile state tracking (local only) |
| `internal/library` | Embedded app recipes and manager |
| `internal/util` | Shared utilities (file ops, string helpers) |

---

## Detailed Component Descriptions

### `internal/cli`

Implements all CLI commands using Cobra. Each command file handles:
- Argument parsing
- Flag handling
- Calling engine functions
- Command-family grouping (for example `app` and `recover` namespaces)
- Formatting output (see [UI Style Guide](../reference/ui-style-guide.md))

### `internal/config`

Loads and validates configuration files:
- `config.yaml` - Global settings
- `profile.yaml` - Profile definitions
- `apps/*.yaml` - App bundle definitions
- **Resolver** - Profile dependency ordering via topological sort

### `internal/engine`

Orchestrates operations by coordinating other packages:
- **Linker** - Dotfile symlink creation with conflict resolution strategies
- **Logger** - Operation logging for rollback support (saved to `.operations/`)
- **HistoryManager** - Historical file snapshot capture and retention in `.history/`
- **Rollback** - Reversal of logged link operations with snapshot restoration
- Profile resolution (includes, conditions)
- Apply/unapply workflows
- State tracking

### `internal/apps`

Manages app bundles:
- **Resolver** - App dependency ordering via topological sort
- Load/save app definitions
- Companion app relationships
- Plugin management
- Auto-detection from paths/commands

### `internal/packages`

Abstract interface over package managers:
- `brew.go` - Homebrew/Linuxbrew
- `apt.go` - Debian/Ubuntu
- `dnf.go` - Fedora/RHEL
- `custom.go` - Custom install scripts

### `internal/shell`

Generates shell integration:
- Combined aliases from all apps
- Function definitions
- Environment variables
- Managed startup/init snippets from app definitions
- Completions loading
- Optional event-based auto-reload hook generation for bash/zsh

### `internal/platform`

Platform abstraction:
- OS detection (macOS, Linux, WSL)
- Distro detection (Ubuntu, Fedora, etc.)
- Path normalization (expand ~, XDG dirs)

### `internal/git`

Git operations:
- Repository initialization
- Clone/pull/push
- Commit with message
- Status checking

### `internal/state`

State management (local only):
- Track applied profiles
- Record app lists per profile
- Timestamp tracking
- State persistence to `~/.gdf/state.yaml`

**Note:** State is LOCAL ONLY and gitignored. It does not sync across machines.
