# GDF Tasks

> **Note**: This file contains only **pending** and **in-progress** tasks.
> Completed tasks are removed and recorded in `CHANGELOG.md`.

---

### Phase 3: Enhanced UX & Advanced Features

#### 3.1 Status, Doctor & Validation
- [ ] Implement rich `gdf status` output
- [ ] Implement `gdf doctor` health check
- [ ] Implement `gdf fix` auto-repair
- [ ] Implement `gdf validate` for YAML checking

#### 3.2 Shell Experience
- [ ] Implement gdf shell completion for bash and zsh
- [ ] Event-based shell auto-reload (using shell hooks to auto-source on prompt)

#### 3.3 Interactive Wizards
- [ ] Implement interactive `gdf setup` wizard
- [ ] Add interactive mode to `gdf add`
- [ ] Add interactive mode to `gdf track`
- [ ] Conflict resolution UI (interactive prompts)

#### 3.4 Advanced Customization
- [ ] Companion apps & plugin support
- [ ] Function management (`gdf fn` commands)
- [ ] Pre/post hooks with error handling options
- [ ] Recipe import/export (`gdf recipe`)
- [ ] First-class Support for Profile Recipes (allow `kind: Profile/v1` in library)
- [ ] Remote Recipe Ecosystem (Git-based Registry & Trust Model)

#### 3.5 App Lifecycle & Cleanup
- [ ] Implement `Uninstall(pkg)` in `packages.Manager` interface
- [ ] Implement `gdf remove --uninstall` to prompt for package removal
- [ ] Implement "Orphaned App" detection (app YAMLs not in any profile)
- [ ] Add `gdf app prune` to remove dangling app definitions
- [ ] Add auto-cleanup prompt when removing the last reference to an app
- [ ] Profile Deletion Strategy: Add `--purge` to `gdf profile delete` to also remove apps only found in that profile
- [ ] Add interactive choice when deleting a profile: (m)igrate to default, (p)urge unique apps, or (l)eave as dangling
- [ ] Implement `Unlink` in engine for clean removal of symlinks

---

### Phase 4: Polish & Future Features

- [ ] Secret management (age encryption)
- [ ] Fish shell support
- [ ] AI-powered recipe generation

---

## Task Management Guidelines

### For AI Agents 

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
