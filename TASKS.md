# GDF Tasks

> **Note**: This file contains only **pending** and **in-progress** tasks.
> Completed tasks are removed and recorded in `CHANGELOG.md`.

---

### Misc tasks

<!-- None pending -->

### Phase 2: User Experience

#### 2.1 UI Foundation (Charmboard & Styles)
- [ ] Evaluate & Integrate Charm libraries (Bubble Tea, Lip Gloss, etc)
- [ ] Create reusable UI components (spinner, list, input, confirm, etc)

#### 2.2 Core App Library (Embedded Recipes)
- [ ] Create library recipe schema
- [ ] Add `mac-preferences` example (package-less bundle)
- [ ] Add 10 core recipes (git, zsh, vim, tmux, etc.)
- [ ] Add 10 development recipes (go, node, python, etc.)
- [ ] Add 10 SRE recipes (kubectl, terraform, aws-cli, etc.)
- [ ] Implement `gdf library list` command
- [ ] Implement `gdf add <app> --from-library`
- [ ] Design Recipe Namespace scheme (local vs core vs remote)
- [ ] Implement Recipe Merging logic (user overrides)

#### 2.3 Interactive Wizards
- [ ] Implement interactive `gdf setup` wizard
- [ ] Add interactive mode to `gdf add`
- [ ] Add interactive mode to `gdf track`

#### 2.4 Status & Doctor
- [ ] Implement rich `gdf status` output
- [ ] Implement `gdf doctor` health check
- [ ] Implement `gdf fix` auto-repair
- [ ] Implement `gdf validate` for YAML checking

#### 2.5 Shell Completions
- [ ] Add --help to provide inline help on available commands, command usage, options, flags etc. 

#### 2.6 Rollback & Cleanup
- [ ] Implement operation logging for all file/package actions
- [ ] Implement `gdf rollback` command (undo last operation)
- [ ] Implement `Unlink` in engine for clean removal of symlinks

#### 2.7 App Lifecycle Management
- [ ] Implement `Uninstall(pkg)` in `packages.Manager` interface
- [ ] Implement `gdf remove --uninstall` to prompt for package removal
- [ ] Implement "Orphaned App" detection (app YAMLs not in any profile)
- [ ] Add `gdf app prune` to remove dangling app definitions
- [ ] Add auto-cleanup prompt when removing the last reference to an app
- [ ] Profile Deletion Strategy: Add `--purge` to `gdf profile delete` to also remove apps only found in that profile
- [ ] Add interactive choice when deleting a profile: (m)igrate to default, (p)urge unique apps, or (l)eave as dangling

---

### Phase 3: Advanced Features

- [ ] Check for new versions (from github) and offer to update.
- [ ] Remote Recipe Ecosystem (Git-based Registry & Trust Model)
- [ ] Companion apps & plugin support
- [ ] Template rendering with documented variables
- [ ] Conditional dotfile linking (evaluate 'when' field)
    - **User Story**: As a GDF user, I want to track OS-specific versions of a dotfile (e.g., `gitconfig.macos`, `gitconfig.linux`) and have them linked to the same local target (e.g., `~/.gitconfig.os`) depending on the current platform.
- [ ] Function management (`gdf fn` commands)
- [ ] Pre/post hooks with error handling options
- [ ] Recipe import/export (`gdf recipe`)
- [ ] Event-based shell auto-reload (using shell hooks to auto-source on prompt)

---

### Phase 4: Polish & AI

- [ ] Conflict resolution UI (interactive prompts)
- [ ] Secret management (age encryption)
- [ ] Fish shell support
- [ ] Import from chezmoi/stow/yadm
- [ ] AI-powered recipe generation

---

## Task Management Guidelines

### For AI Agents (Gemini/Antigravity)

1. **Complete BLOCKER section first** - No Phase 1 coding until done
2. **Pick tasks from top to bottom** within each section
3. **Mark tasks as in-progress** by changing `[ ]` to `[/]`
4. **Remove completed tasks** from this file
5. **Add completed work to `CHANGELOG.md`** under Unreleased
6. **Follow TDD**: Write tests before implementation
7. **Update docs** when adding new features

### Task States

- `[ ]` - Pending (not started)
- `[/]` - In Progress (actively being worked on)
- Completed tasks are **removed** and logged in CHANGELOG.md
