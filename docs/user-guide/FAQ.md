# GDF FAQ

Common scenarios, expected behavior, and safe recovery steps.

## Setup and First Apply

### 1) `gdf apply` fails because target files already exist. What should I do?

This is controlled by `conflict_resolution.dotfiles` in `~/.gdf/config.yaml`.

- `error`: fail safely when a target already exists.
- `backup_and_replace`: move the existing file to `<target>.gdf.bak` and link managed source.
- `replace`: replace target directly (with history snapshot capture for rollback).

Recommended first-run flow:

```bash
gdf apply --dry-run <profile>
```

Then choose a conflict strategy explicitly in config before live apply.

### 2) I forgot to source shell integration. How do I fix it?

Load current shell session:

```bash
source ~/.gdf/generated/init.sh
```

If missing in RC file, add:

```bash
[ -f ~/.gdf/generated/init.sh ] && source ~/.gdf/generated/init.sh
```

Then reload RC (`source ~/.zshrc` or `source ~/.bashrc`).

### 3) `gdf app add` succeeded but nothing changed on my machine.

`gdf app add` updates desired state in your repo. It does not mutate live targets unless you run apply.

```bash
gdf apply <profile>
```

For guarded immediate activation:

```bash
gdf app add <app> -p <profile> --apply
```

### 4) I already have many existing dotfiles. Use `track` or `import`?

- Use `gdf app track` for one-off adoption of a specific file.
- Use `gdf app import --preview` first for bulk discovery and guided mapping.

## Multi-Profile and Selection

### 5) Why am I being prompted to choose a profile?

If `--profile` is omitted, profile-dependent commands resolve as:

- no profiles: error
- one profile: auto-select
- multiple profiles: prompt

Use `-p/--profile` to avoid interactive selection.

### 6) Which profile should I apply on a new machine?

Use your base/shared profile first, then context-specific profiles. Always preview first:

```bash
gdf apply --dry-run base
gdf apply base
```

### 7) I applied one profile but expected included profiles/apps too.

Profiles with `includes` are resolved during apply. Validate what is included:

```bash
gdf profile show <profile>
```

## Recovery and Safety

### 8) I accidentally overwrote or removed managed links.

Use rollback on latest operations:

```bash
gdf recover rollback
```

For one target:

```bash
gdf recover rollback --target ~/.gitconfig --choose-snapshot
```

### 9) How do I recover one file without undoing everything?

Use targeted rollback with snapshot choice:

```bash
gdf recover rollback --target <path> --choose-snapshot
```

### 10) What is the difference between `recover rollback` and `recover restore`?

- `recover rollback`: undo recent logged operations and restore from snapshots.
- `recover restore`: offboarding flow; replace managed symlinks with real files at original targets and restore shell alias sourcing away from GDF-managed generation.

### 11) How much history is kept? Why do old snapshots disappear?

History is quota-based (`history.max_size_mb` in `~/.gdf/config.yaml`). Older snapshots are evicted when quota is exceeded.

## Sync and Multi-Machine

### 12) I synced from another machine and now have drift.

Inspect drift before applying:

```bash
gdf status
gdf status diff
```

Then preview:

```bash
gdf apply --dry-run <profile>
```

### 13) Can I safely pull/sync while local changes exist?

Recommended sequence:

```bash
gdf status
gdf save "local updates"
gdf sync
gdf apply --dry-run <profile>
```

Resolve git conflicts inside `~/.gdf` before apply.

### 14) Why does applied state differ across machines if repo is the same?

`~/.gdf/state.yaml` is local-only and not synced. It affects automatic profile reuse when `gdf apply` runs without profile args.

## App Lifecycle and Cleanup

### 15) Remove app from profile vs uninstall app from system

- `gdf app remove <app> -p <profile>`: remove from desired profile only.
- `gdf app remove <app> -p <profile> --uninstall`: also unlink managed dotfiles and attempt package uninstall when safe.

Preview removal first:

```bash
gdf app remove <app> -p <profile> --uninstall --dry-run
```

### 16) What are orphaned apps? Should I prune them?

Orphaned apps are local app definitions not referenced by any profile.

```bash
gdf app prune --dry-run
gdf app prune
```

Default prune archives orphans instead of deleting them.

### 17) I removed an app and lost aliases/functions.

Shell artifacts are generated from current profile/app definitions at apply time. Re-add desired app/shell definitions and re-apply.

## Troubleshooting and Automation

### 18) Dry-run passes, but apply fails.

Common causes:

- package manager permissions/availability
- missing source files in `~/.gdf/dotfiles/...`
- target conflicts under stricter conflict mode
- shell or filesystem permission issues

Check:

```bash
gdf health validate
gdf health doctor
gdf status diff
```

### 19) Non-interactive/CI runs fail due to prompts.

Use deterministic flags:

- `--non-interactive`
- `--yes` (where supported)
- `--json` for machine-readable parsing

### 20) `gdf health doctor` reports issues I do not understand.

Run doctor, then preview fixes:

```bash
gdf health doctor
gdf health fix --dry-run
```

Use guarded mode for higher-impact backup-before-write fixes:

```bash
gdf health fix --guarded --dry-run
```

## Scenario: Apply on a Machine that Already Has Local Dotfiles

Example: your machine already has `~/.gitconfig`, and synced profile includes `git`.

Behavior depends on `conflict_resolution.dotfiles`:

- `error`: apply stops for that target, existing file remains unchanged.
- `backup_and_replace`: existing file moved to `~/.gitconfig.gdf.bak` (with rotation).
- `replace`: existing file replaced.

In replacement modes, GDF also captures history snapshots for rollback.

Recover old content:

1. From `.gdf.bak` files (if using backup mode).
2. From snapshots:

```bash
gdf recover rollback --target ~/.gitconfig --choose-snapshot
```

## Scenario: Restore Dotfiles Before Removing GDF (Without Uninstalling Apps)

Use this when you want to stop GDF management but keep installed software.

1. Restore managed links back to regular files:

```bash
gdf recover restore
```

2. Remove apps from profiles without uninstalling packages (repeat per app/profile):

```bash
gdf app remove <app> -p <profile>
```

Do not use `--uninstall` for this goal.

3. Cleanup orphaned app definitions:

```bash
gdf app prune --dry-run
gdf app prune
```

4. Optional: archive the whole repo for clean re-init later:

```bash
mv ~/.gdf ~/.gdf.backup-$(date +%Y%m%d-%H%M%S)
```
