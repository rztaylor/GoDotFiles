# Architecture Review: Edge Cases & Design Decisions

> **Status:** ✅ Finalized — Ready for Phase 1 implementation.

This document identifies gaps and edge cases that require architectural decisions.

---

## ✅ Finalized Decisions

### 1. Conflict Resolution

**Problem:** What happens when two profiles define the same alias or dotfile differently?

**Decision:** Use strategy-based resolution with sensible defaults.

| Conflict Type | Options | Recommended |
|---------------|---------|-------------|
| Same alias, different command | Last wins / Error / Prompt | Last wins + warn |
| Same dotfile path | Error / Backup + override | Error (require explicit) |
| Same app in multiple profiles | Merge / Override | Merge settings |

**Config support:**
```yaml
# config.yaml
conflict_resolution:
  aliases: last_wins    # or: error, prompt
  dotfiles: error       # or: backup_and_replace, prompt
```

---

### 2. Rollback and Transaction Safety

**Problem:** No recovery if `apply` fails midway.

**Decision:** Implement operation logging for rollback.

```
~/.gdf/.operations/
├── 2024-02-08T21-30-00.log  # What was changed
└── state.backup.yaml        # Pre-operation state

~/.gdf/.history/
└── <snapshot-id>.snap       # Historical file copy used for rollback
```

**Commands:**
- `gdf rollback` - Undo last operation
- `gdf apply --dry-run` - Preview changes (MVP)

---

### 3. App Dependencies

**Problem:** Can't express "kubectl-neat requires kubectl + krew".

**Decision:** Add `dependencies` field to AppBundle schema.

```yaml
# apps/kubectl-neat.yaml
name: kubectl-neat
dependencies:
  - kubectl
  - krew
```

Engine performs topological sort during apply.

---

### 4. Secrets and Sensitive Data

**Problem:** SSH keys, API tokens shouldn't be in plain git.

**Decision:** Support `secret: true` flag on dotfiles.

```yaml
dotfiles:
  - source: aws/credentials
    target: ~/.aws/credentials
    secret: true    # Encrypted with age, gitignored
```

Phase 1: Document as limitation, recommend gitignore
Phase 2+: Implement age encryption

---

### 5. State Synchronization

**Problem:** Is `state.yaml` local or shared via git?

**Decision:** state.yaml is local-only (gitignored).

- `state.yaml` tracks what's applied on THIS machine
- Not committed to git (machine-specific)
- Profile/app definitions ARE committed

```
~/.gdf/
├── .gitignore         # Contains: state.yaml, .operations/, .history/
├── state.yaml         # Local: {applied_profiles: [base, sre]}
└── profiles/          # Shared via git
```

---

### 6. Managed Shell Startup Tasks

**Problem:** Many apps require manual RC/profile edits (PATH/eval/source lines), which creates unmanaged drift and makes app removal incomplete.

**Decision:** Keep a single RC source line and generate startup snippets from app YAML into `~/.gdf/generated/init.sh`.

```yaml
shell:
  init:
    - name: fnm-env
      bash: eval "$(fnm env --shell bash)"
      zsh: eval "$(fnm env --shell zsh)"
      guard: command -v fnm >/dev/null 2>&1
```

**Rationale:**
- Centralized ownership and deterministic regeneration during `gdf apply`
- No per-app mutation of user RC files
- Add/remove app automatically adds/removes startup behavior
- Extensible to future shells without changing RC integration model

---

## 🟡 App Bundles Without Packages

**Insight:** An app bundle should NOT require a package installation.

**Use cases:**
- `mac-preferences` - Just runs `defaults write` commands
- `shell-aliases` - Only defines aliases, no package
- `env-vars` - Only sets environment variables
- `git-config` - Only manages dotfiles

**Schema update:**

```yaml
# apps/mac-preferences.yaml
name: mac-preferences
description: macOS system preferences via defaults

# No package field = no installation step

# Platform-specific hooks
hooks:
  apply:
    - when: os == 'macos'
      run: |
        # Dock settings
        defaults write com.apple.dock autohide -bool true
        defaults write com.apple.dock tilesize -int 48
        
        # Finder settings
        defaults write com.apple.finder ShowPathbar -bool true
        
        # Restart affected apps
        killall Dock Finder
```

This makes app bundles a flexible container for:
- Packages + dotfiles + aliases (traditional)
- Just dotfiles (config-only)
- Just commands (system preferences)
- Just aliases/functions (shell customization)

---

## Medium Priority Issues

| Issue | Decision |
|-------|----------|
| Partial apply | Add `--only`, `--skip` flags |
| Template variables | Document available vars: `{{.OS}}`, `{{.Hostname}}`, etc. |
| Binary configs | Mark as unsupported initially |
| Hook failures | Fail by default, add `optional: true` flag |
| Validation | Implement `gdf validate` command |

---

## Architecture Decisions Log

| Decision | Choice | Rationale |
|----------|--------|-----------|
| state.yaml | Local only | Machine-specific state shouldn't sync |
| Conflicts | Last wins + warn | Predictable, user sees warning |
| Dependencies | Explicit field | Required for complex tools |
| Secrets | Flag + gitignore (Phase 1) | Encryption deferred |
| Package-less bundles | Supported | Enables mac-preferences use case |
| Shell startup tasks | Generated in init.sh | Keep RC files clean and app-scoped |

---

## Related Documentation

- [YAML Schemas](../reference/yaml-schemas.md)
- [Architecture Overview](overview.md)
