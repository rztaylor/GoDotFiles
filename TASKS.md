# GDF Tasks

> **Note**: This file contains only **pending** and **in-progress** tasks.
> Completed tasks are removed and recorded in `CHANGELOG.md`.

---

## Priority 0: Reliability, Safety, and Trust (Do First)

### 0.4 Deferred Follow-ups From Initial Priority 0 Scope
- [ ] Expand `gdf status diff` with optional full patch-style diffs and clear performance limits
- [ ] Expand `gdf health fix` beyond safe-only remediations with explicit guarded modes for higher-impact fixes (with previews and backups)
- [ ] Expand `--non-interactive` and `--yes` prompt controls to all interactive command paths (for example `init`, `app add`, and `app install` prompts)
- [ ] Improve drift/diff performance with incremental scanning or cached metadata to keep large repos responsive

---

## Priority 1: Safe App Lifecycle and Cleanup

### 1.1 Install/Uninstall Symmetry
- [ ] Implement `Uninstall(pkg)` in `packages.Manager` interface
- [ ] Implement `gdf app remove --uninstall` with explicit confirmation and preview
- [ ] Implement `Unlink` in engine for clean symlink removal and rollback-safe behavior

### 1.2 Orphan and Dangling State Management
- [ ] Implement orphaned app detection (app YAMLs not referenced by any profile)
- [ ] Add `gdf app prune` to remove or archive dangling app definitions
- [ ] Add auto-cleanup prompt when removing the last profile reference to an app

### 1.3 Profile Deletion Safety
- [ ] Add `--purge` to `gdf profile delete` for removing apps unique to that profile
- [ ] Add interactive delete strategy: migrate to `default`, purge unique apps, or leave as dangling
- [ ] Provide dry-run preview for profile deletion impact across apps/dotfiles/packages

---

## Priority 2: Onboarding, Migration, and Day-1 Adoption

### 2.1 Guided Setup
- [ ] Implement interactive setup as `gdf init setup` (or equivalent grouped flow) for first-run profile/app/bootstrap
- [ ] Add conflict-resolution UI with explicit, auditable decisions

### 2.2 Migration from Existing Environments
- [ ] Implement `gdf app import` to discover and adopt existing dotfiles, aliases, and common tool configs
- [ ] Add import modes: preview-only, guided mapping, and apply
- [ ] Add secret-aware import flow (detect likely sensitive paths and require explicit handling choice)

### 2.3 Interactive Authoring
- [ ] Add interactive mode to `gdf app add` with recipe suggestions and dependency awareness
- [ ] Add interactive mode to `gdf app track` with target/path conflict previews

---

## Priority 3: Reproducibility, Secrets, and Policy Controls

### 3.1 Reproducible Environments
- [ ] Implement lock file support for resolved package sources/versions per platform
- [ ] Add lock refresh and verification workflow integrated with `apply` and CI checks

### 3.2 Secrets Management (Beyond Encryption Primitive)
- [ ] Implement secret management workflow (encrypt/decrypt/edit) with `age`
- [ ] Add secret templates/placeholders and missing-secret preflight checks
- [ ] Add secret rotation and re-encryption workflow

### 3.3 Trust and Governance
- [ ] Define policy controls for external/remote recipes (allowlists, pinning, provenance expectations)
- [ ] Add trust policy enforcement during recipe resolution and apply-time execution

---

## Priority 4: Recipe Ecosystem and Extensibility

### 4.1 Recipe Model Expansion
- [ ] Implement first-class support for profile recipes (`kind: Profile/v1` in library)
- [ ] Implement recipe import/export (`gdf recipe`) with schema/version validation

### 4.2 Remote Recipe Ecosystem
- [ ] Implement Git-based remote recipe registry workflows
- [ ] Add recipe source pinning and update/upgrade UX
- [ ] Add safe defaults for remote changes (review-before-apply flow)

### 4.3 Plugin and Bundle Extensibility
- [ ] Implement companion apps and plugin support with explicit compatibility checks
- [ ] Implement pre/post hooks with policy-aware error handling options
- [ ] Implement function management under grouped command namespace (for example `gdf shell fn ...`)

---

## Priority 5: Platform and Long-Horizon Enhancements

- [ ] Add fish shell support (parity target: init integration, reload behavior, completion guidance)
- [ ] Add AI-assisted recipe generation (strictly opt-in, with validation and safety review workflow)

---

## Task Management Guidelines

### For AI Agents

1. **Execute tasks from top to bottom** (Priority -1, then Priority 0 to Priority 5)
2. **Mark tasks as in-progress** by changing `[ ]` to `[/]`
3. **Remove completed tasks** from this file
4. **Add completed work to `CHANGELOG.md`** under Unreleased
5. **Follow TDD**: Write tests before implementation
6. **Update docs** when adding new features

### Task States

- `[ ]` - Pending (not started)
- `[/]` - In Progress (actively being worked on)
- Completed tasks are **removed** and logged in CHANGELOG.md
