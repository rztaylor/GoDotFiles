# YAML Schema Reference

This document defines the YAML schemas for gdf configuration files.

---

## v1 Stability Contract

For the public `1.x` series, these schema guarantees are stable:
- Core kinds remain `App/v1`, `Profile/v1`, and `Config/v1`.
- Existing `v1` fields keep current meaning; new fields may be additive and optional only.
- `conflict_resolution` defaults remain:
  - `aliases: last_wins`
  - `dotfiles: error`
- Missing or invalid `kind` values continue to be rejected.

Breaking reinterpretation of `v1` fields is deferred to a future major version.

---

## App Bundle Schema (`apps/*.yaml`)

```yaml
kind: App/v1              # Required: resource type and version
name: string              # Required: app identifier
description: string       # Optional: human-readable description

# ─────────────────────────────────────────────────────────────────
# DEPENDENCIES
# ─────────────────────────────────────────────────────────────────
dependencies:             # Optional: apps that must be installed first
  - string                # App names (topologically sorted during apply)

# ─────────────────────────────────────────────────────────────────
# PACKAGE INSTALLATION (optional - omit for package-less bundles)
# ─────────────────────────────────────────────────────────────────
package:
  # Simple form: package name per manager
  brew: string            # Homebrew/Linuxbrew package name
  apt: string             # Debian/Ubuntu package name
  dnf: string             # Fedora/RHEL package name
  pacman: string          # Arch Linux package name
  
  # Extended form: with repository configuration
  apt:
    name: string          # Package name
    repo: string          # APT repository URL (optional)
    key: string           # GPG key URL (optional)
    
  # Custom installation script
  custom:
    script: string        # Shell script to run
    sudo: boolean         # Whether script requires sudo (default: false)
    confirm: boolean      # Require user confirmation (default: true for scripts)
    
  # Conditions for which installer to use
  prefer:
    macos: brew           # Use brew on macOS
    linux: apt            # Use apt on Linux (if available)
    wsl: apt              # Use apt on WSL

# ─────────────────────────────────────────────────────────────────
# DOTFILE MANAGEMENT
# ─────────────────────────────────────────────────────────────────
dotfiles:
  # Simple form: same path on all platforms
  - source: string        # Path in ~/.gdf/dotfiles/ (relative)
    target: string        # Target path (~ expanded)
    
  # Platform-specific targets
  - source: string
    target:
      default: string     # Default target path
      macos: string       # macOS-specific path
      linux: string       # Linux-specific path
      wsl: string         # WSL-specific path (often /mnt/c/...)
      
  # Conditional dotfiles
  - source: string
    target: string
    when: string          # Condition expression:
                          # "os == 'macos'"
                          # "hostname =~ '^work-.*'"
                          # "os == 'linux' OR os == 'wsl'"
                          # "(os == 'linux' OR os == 'wsl') AND arch == 'amd64'"
    
  # Template rendering
  - source: string
    target: string
    template: boolean     # Render Go templates (default: false)
    
  # Secret files (gitignored, not committed)
  - source: string
    target: string
    secret: boolean       # If true, add to .gitignore and warn (default: false)

# ─────────────────────────────────────────────────────────────────
# SHELL INTEGRATION
# ─────────────────────────────────────────────────────────────────
shell:
  aliases:
    name: string          # alias name → command
    
  functions:
    name: |               # Function body (multiline)
      function_name() {
        ...
      }
      
  env:
    VAR_NAME: string      # Environment variable
    
  completions:
    bash: string          # Command to generate bash completions (captured during gdf apply)
    zsh: string           # Command to generate zsh completions (captured during gdf apply)

  init:
    - name: string        # Required: unique snippet id within this app
      common: string      # Optional: default command for all shells
      bash: string        # Optional: bash-specific command (overrides common)
      zsh: string         # Optional: zsh-specific command (overrides common)
      guard: string       # Optional: condition wrapper (if <guard>; then ...)
                          # At least one of common/bash/zsh is required

# ─────────────────────────────────────────────────────────────────
# HOOKS
# ─────────────────────────────────────────────────────────────────
hooks:
  pre_install:            # Run before package installation
    - string              # Shell commands
  post_install:           # Run after package installation
    - string
  pre_link:               # Run before dotfile linking
    - string
  post_link:              # Run after dotfile linking
    - string
  apply:                  # Run during apply (for package-less bundles)
    - run: string         # Shell commands to run
      when: string        # Optional condition (e.g., "os == 'macos'")

# ─────────────────────────────────────────────────────────────────
# COMPANIONS & PLUGINS
# ─────────────────────────────────────────────────────────────────
companions:
  - string                # Related apps to suggest

plugins:
  - name: string          # Plugin identifier
    install: string       # Install command (e.g., "kubectl krew install neat")
```

---

## Platform-Specific Examples

### Azure CLI - Different Install Methods

```yaml
kind: App/v1
name: azure-cli
description: Azure command-line interface

package:
  brew: azure-cli
  
  # apt with Microsoft's repository
  apt:
    name: azure-cli
    repo: https://packages.microsoft.com/repos/azure-cli/
    key: https://packages.microsoft.com/keys/microsoft.asc
    
  # Fallback custom script (requires confirmation)
  custom:
    script: curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
    sudo: true
    confirm: true         # ALWAYS prompts user before running

dotfiles:
  - source: azure/config
    target:
      default: ~/.azure/config
      wsl: /mnt/c/Users/${WINDOWS_USER}/.azure/config
```

### MySQL Client - OS-Specific Paths

```yaml
kind: App/v1
name: mysql-client
description: MySQL command-line client

package:
  brew: mysql-client
  apt: mysql-client
  dnf: mysql

dotfiles:
  - source: my.cnf
    target:
      macos: ~/.my.cnf
      linux: ~/.my.cnf
      wsl: ~/.my.cnf       # Or /mnt/c/my.ini for Windows MySQL
```

### kubectl - Conditional Config

```yaml
kind: App/v1
name: kubectl
description: Kubernetes CLI

package:
  brew: kubectl
  apt: kubectl
  dnf: kubernetes-client

dotfiles:
  # Standard kubeconfig
  - source: kube/config
    target: ~/.kube/config
    
  # Work-specific config (only on work machines)
  - source: kube/config.work
    target: ~/.kube/config.d/work
    when: hostname =~ '^work-.*'

  # Linux + WSL shared config
  - source: kube/config.posix
    target: ~/.kube/config.d/posix
    when: os == 'linux' OR os == 'wsl'
```

---

### fnm - Managed Shell Startup

```yaml
kind: App/v1
name: fnm
description: Fast Node.js version manager

package:
  brew: fnm
  apt:
    name: fnm

shell:
  init:
    - name: fnm-path
      common: export PATH="$HOME/.local/share/fnm:$PATH"
    - name: fnm-env
      bash: eval "$(fnm env --use-on-cd --shell bash)"
      zsh: eval "$(fnm env --use-on-cd --shell zsh)"
      guard: command -v fnm >/dev/null 2>&1
```

---

## Security Model

### Custom Scripts

Custom installation scripts (piped bash, curl | bash, etc.) have security implications:

```yaml
package:
  custom:
    script: curl -sL https://example.com/install.sh | bash
    sudo: true          # Script needs root privileges
    confirm: true       # User MUST confirm before execution
```

**gdf behavior:**
1. **Always warns** about custom scripts before execution
2. **Shows script source URL** for inspection
3. **Requires explicit confirmation** (cannot be bypassed with --yes)
4. **Logs all custom script executions** for audit
5. **Scans hook/custom commands for high-risk patterns** (for example `curl|wget` piped to shell) and prompts during `gdf apply` unless `--allow-risky` is used

Example interaction:
```
$ gdf apply sre

⚠️  Custom install script for 'azure-cli':
    Source: https://aka.ms/InstallAzureCLIDeb
    Requires: sudo
    
    This will download and execute a remote script.
    Review the script before proceeding: gdf show-script azure-cli
    
? Proceed with script execution? [y/N] 
```

### Sudo Handling

```yaml
package:
  apt: some-package        # apt install typically needs sudo
  
  custom:
    script: ...
    sudo: true             # Explicitly marked as needing sudo
```

**gdf behavior:**
- Package managers (apt, dnf) handle sudo themselves
- Custom scripts with `sudo: true` will run with sudo prefix
- User is informed when sudo is required
- Never runs sudo without user awareness

---

## Profile Schema (`profile.yaml`)

```yaml
kind: Profile/v1          # Required
name: string              # Required: profile identifier
description: string       # Optional: description
includes:                 # Include other profiles
  - string
  
conditions:               # Conditional includes
  - if: string            # Condition expression
    include_apps:
      - string
    exclude_apps:
      - string
```

### Condition Expressions

| Variable | Values | Example |
|----------|--------|---------|
| `os` | macos, linux, wsl | `os == 'wsl'` |
| `distro` | ubuntu, debian, fedora, arch, etc. | `distro == 'ubuntu'` |
| `hostname` | machine hostname | `hostname =~ '^work-.*'` |
| `arch` | amd64, arm64 | `arch == 'arm64'` |

Supported operators: `==`, `!=`, `=~` (regex match).

---

## Global Aliases Schema (`aliases.yaml`)

Stores aliases not tied to any app bundle. These are created when `gdf alias add` is used for commands that don't match an existing app (e.g., `ls`, `cd`, `cat`).

```yaml
aliases:
  ll: "ls -la"              # Simple command alias
  la: "ls -A"               # Another simple alias
  "..": "cd .."             # Shell navigation shortcut
  "...": "cd ../.."         # Quoted keys for special characters
  cls: clear                # Clear screen
```

| Field | Type | Description |
|-------|------|-------------|
| `aliases` | `map[string]string` | Map of alias name → command |

**Notes:**
- This file is auto-created on first `gdf alias add` for an unmatched command
- App-bundle aliases take precedence over global aliases with the same name during shell generation
- Use `gdf alias list` to see both app and global aliases
- Use `gdf alias remove <name>` to remove from either location

---

## Global Config Schema (`config.yaml`)

```yaml
kind: Config/v1           # Required
# Repository settings
git:
  remote: string          # Git remote URL
  branch: string          # Default branch (default: main)
  
# Default shell
shell: zsh | bash | fish

# Conflict resolution strategy
conflict_resolution:
  aliases: last_wins | error | prompt    # Default: last_wins (+ warning)
  dotfiles: error | backup_and_replace | prompt  # Default: error

# Package manager preferences
package_manager:
  prefer:
    macos: brew
    linux: apt            # or dnf, pacman
    wsl: apt              # optional WSL-specific override
    
# Security settings
security:
  confirm_scripts: true   # Always confirm custom scripts (default: true)
  log_scripts: true       # Log script executions (default: true)

# Snapshot history retention
history:
  max_size_mb: 512        # Max size for ~/.gdf/.history (default: 512)

# Shell integration behavior
shell_integration:
  auto_reload_enabled: true | false   # Default: false (recommended true for interactive shells)
```

Preference precedence during `gdf apply`:
1. `package.prefer` in app bundle
2. `package_manager.prefer` in global config
3. platform auto-detection

---

## State Schema (`state.yaml`)

> [!IMPORTANT]
> This file is **LOCAL ONLY** and automatically gitignored. State does not sync across machines.

```yaml
kind: State/v1
# Applied profile tracking
applied_profiles:
  - name: string              # Profile name
    apps: []string            # Apps in this profile
    applied_at: timestamp     # When profile was applied (RFC3339 format)
    
# Last apply operation timestamp
last_applied: timestamp       # When any profile was last applied (RFC3339 format)
```

### Example

```yaml
kind: State/v1
applied_profiles:
  - name: base
    apps:
      - git
      - zsh
      - vim
    applied_at: 2024-02-11T18:30:00Z
  - name: work
    apps:
      - kubectl
      - terraform
    applied_at: 2024-02-11T18:31:00Z
    
last_applied: 2024-02-11T18:31:00Z
```

### Purpose

The state file tracks which profiles have been applied to the current machine. This enables:
- `gdf status` to show what's currently active
- Future rollback functionality
- Tracking of when profiles were last applied

### Location

- **Path**: `~/.gdf/state.yaml`
- **Visibility**: Local only (gitignored)
- **Management**: Automatically updated by `gdf apply`
