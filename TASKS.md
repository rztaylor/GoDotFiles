# GDF Tasks (1.0-Focused Roadmap)

> **Note**: This file contains only **pending** and **in-progress** tasks.
> Completed tasks are removed and recorded in `CHANGELOG.md`.

---

## Release Objective

Ship a public `1.0.0` that is safe, predictable, and stable to build on.

For 1.0 prioritization:
- Prevent data loss and unsafe destructive operations.
- Make automation behavior deterministic (`--non-interactive`, `--yes`, exit codes, JSON).
- Freeze core user workflows so later versions add capabilities without forcing usage changes.

---

## Priority 0: 1.0 Release Blockers (Must Ship Before Public 1.0)

No remaining pending tasks.

---

## Priority 1: 1.0.x Hardening (Important, Backward-Compatible)

### 1.1 Lifecycle Hygiene and Repository Cleanliness
- [ ] Implement orphaned app detection (app YAMLs not referenced by any profile)
- [ ] Add `gdf app prune` to remove or archive dangling app definitions
- [ ] Add optional cleanup guidance when removing the last profile reference to an app (must respect non-interactive behavior)

### 1.2 Profile UX Improvements
- [ ] Add interactive delete strategy: migrate to `default`, purge unique apps, or leave as dangling

### 1.3 Drift Observability and Scale
- [ ] Expand `gdf status diff` with optional full patch-style diffs and clear performance limits
- [ ] Improve drift/diff performance with incremental scanning or cached metadata to keep large repos responsive

### 1.4 Health Fix Expansion (Guarded)
- [ ] Expand `gdf health fix` beyond safe-only remediations with explicit guarded modes for higher-impact fixes (with previews and backups)

---

## Priority 2: Post-1.0 Adoption and Reproducibility (1.1+)

### 2.1 Onboarding and Migration
- [ ] Implement interactive setup as `gdf init setup` (or equivalent grouped flow) for first-run profile/app/bootstrap
- [ ] Add conflict-resolution UI with explicit, auditable decisions
- [ ] Implement `gdf app import` to discover and adopt existing dotfiles, aliases, and common tool configs
- [ ] Add import modes: preview-only, guided mapping, and apply
- [ ] Add secret-aware import flow (detect likely sensitive paths and require explicit handling choice)

### 2.2 Interactive Authoring
- [ ] Add interactive mode to `gdf app add` with recipe suggestions and dependency awareness
- [ ] Add interactive mode to `gdf app track` with target/path conflict previews

### 2.3 Reproducible Environments and Secrets
- [ ] Implement lock file support for resolved package sources/versions per platform
- [ ] Add lock refresh and verification workflow integrated with `apply` and CI checks
- [ ] Implement secret management workflow (encrypt/decrypt/edit) with `age`
- [ ] Add secret templates/placeholders and missing-secret preflight checks
- [ ] Add secret rotation and re-encryption workflow

### 2.4 Trust and Governance for Remote Content
- [ ] Define policy controls for external/remote recipes (allowlists, pinning, provenance expectations)
- [ ] Add trust policy enforcement during recipe resolution and apply-time execution

---

## Priority 3: Ecosystem and Extensibility (Later)

### 3.1 Recipe Ecosystem
- [ ] Implement first-class support for profile recipes (`kind: Profile/v1` in library)
- [ ] Implement recipe import/export (`gdf recipe`) with schema/version validation
- [ ] Implement Git-based remote recipe registry workflows
- [ ] Add recipe source pinning and update/upgrade UX
- [ ] Add safe defaults for remote changes (review-before-apply flow)

### 3.2 Plugin and Bundle Extensibility
- [ ] Implement companion apps and plugin support with explicit compatibility checks
- [ ] Implement pre/post hooks with policy-aware error handling options
- [ ] Implement function management under grouped command namespace (for example `gdf shell fn ...`)

---

## Priority 4: Platform and Long-Horizon Experiments

- [ ] Add fish shell support (parity target: init integration, reload behavior, completion guidance)
- [ ] Add AI-assisted recipe generation (strictly opt-in, with validation and safety review workflow)

---

## Task Management Guidelines

### For AI Agents

1. **Execute tasks from top to bottom** (Priority 0 to Priority 4)
2. **Mark tasks as in-progress** by changing `[ ]` to `[/]`
3. **Remove completed tasks** from this file
4. **Add completed work to `CHANGELOG.md`** under Unreleased
5. **Follow TDD**: Write tests before implementation
6. **Update docs** when adding new features

### Task States

- `[ ]` - Pending (not started)
- `[/]` - In Progress (actively being worked on)
- Completed tasks are **removed** and logged in CHANGELOG.md
