# Architecture Overview

GDF is built with a layered architecture that separates concerns and enables cross-platform support.

## High-Level Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                         CLI (cobra)                              │
│  init, add, track, alias, profile, apply, status, health, sync   │
└─────────────────────────────┬────────────────────────────────────┘
                              │
┌─────────────────────────────▼────────────────────────────────────┐
│                        App Engine                                │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│  │ App Manager │ │   Profile   │ │  Dotfile    │ │    Shell    │ │
│  │ (bundles)   │ │  Resolver   │ │  Linker     │ │  Generator  │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────────┐ │
│  │   Library   │ │ Smart       │ │      Package Installer      │ │
│  │  (recipes)  │ │ Detection   │ │  (brew, apt, dnf, custom)   │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────────┘ │
└─────────────────────────────┬────────────────────────────────────┘
                              │
┌─────────────────────────────▼────────────────────────────────────┐
│                    Platform Layer                                │
│         OS/distro detection, path resolution, sudo handling      │
└─────────────────────────────┬────────────────────────────────────┘
                              │
┌─────────────────────────────▼────────────────────────────────────┐
│                    Git Storage Backend                           │
└──────────────────────────────────────────────────────────────────┘
```

## Design Principles

1. **App Bundle as First-Class Concept** - Packages, dotfiles, and shell integration belong together
2. **80/20 CLI** - Common operations via CLI, advanced via YAML
3. **Cross-Platform Abstraction** - OS differences hidden behind interfaces
4. **Git-Native Storage** - Everything is version controlled
5. **Security-Conscious** - User confirmation for privileged operations

---

## Cross-Platform Support

A single app bundle works across macOS, Linux, and WSL. The platform layer resolves differences automatically.

### Platform-Aware Packages

Each app specifies packages per package manager:

```yaml
package:
  brew: kubectl        # macOS, Linuxbrew
  apt: kubectl         # Debian/Ubuntu
  dnf: kubernetes-client  # Fedora/RHEL
```

The engine selects the appropriate installer based on detected OS/distro.

### Platform-Aware Dotfiles

Dotfiles can target different paths per OS:

```yaml
dotfiles:
  - source: azure/config
    target:
      default: ~/.azure/config
      wsl: /mnt/c/Users/${WINDOWS_USER}/.azure/config
```

### Conditional Inclusion

Both profiles and individual dotfiles support conditions:

```yaml
# Profile level
conditions:
  - if: os == 'macos'
    include_apps: [iterm2]

# Dotfile level
dotfiles:
  - source: work.conf
    target: ~/.config/work.conf
    when: hostname =~ '^work-.*'
```

**See**: [YAML Schemas](../reference/yaml-schemas.md) for complete syntax.

---

## Security Model

### Sudo Handling

- Standard package managers (apt, dnf) invoke sudo when needed
- Custom scripts explicitly declare sudo requirements: `sudo: true`
- User is always informed when sudo is required

### Custom Script Safety

Piped bash scripts (`curl | bash`) require explicit user confirmation:

```yaml
package:
  custom:
    script: curl -sL https://example.com/install.sh | bash
    sudo: true
    confirm: true   # Cannot be bypassed
```

**gdf behavior:**
1. Displays warning with script source URL
2. Offers to show script contents: `gdf show-script <app>`
3. Requires explicit `y` confirmation (no default)
4. Logs all script executions for audit

### Configuration Analysis

Before applying any changes, GDF scans the configuration for potential security risks:

- **Malicious Pattern Detection**: Identifies risky install patterns (e.g., `curl | sh`)
- **Hook Inspection**: Flags pre/post hooks that execute arbitrary shell commands
- **Review Prompt**: If risks are detected, the user is warned and offered a chance to review the relevant content before proceeding

---

## Package Dependencies

```
cmd/gdf/main.go
    └── internal/cli
            ├── internal/engine
            │       ├── internal/apps
            │       ├── internal/config
            │       ├── internal/packages
            │       ├── internal/shell
            │       └── internal/platform
            └── internal/git
```

## Related Documentation

- [Component Details](components.md)
- [Versioning Strategy](versioning.md)
- [UI Style Guide](../reference/ui-style-guide.md)
- [YAML Schemas](../reference/yaml-schemas.md)
- [CLI Reference](../reference/cli.md)
