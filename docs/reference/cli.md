# CLI Reference

Complete reference for all gdf commands.

> [!TIP]
> All CLI output and interactive components should follow the [UI Style Guide](ui-style-guide.md).

## Global Flags

| Flag            | Description               |
| --------------- | ------------------------- |
| `-v, --verbose` | Enable verbose output     |
| `--yes`         | Auto-approve supported prompts |
| `--non-interactive` | Disable supported prompts and fail when confirmation is required |
| `-h, --help`    | Show help for any command |

## CLI Information Architecture

### Command taxonomy

- Keep high-frequency workflows as top-level commands: `init`, `apply`, `status`, `save`, `push`, `pull`, `sync`.
- Group domain-specific lifecycle operations under command families:
  - `gdf app ...` for app bundle and recipe workflows
  - `gdf recover ...` for rollback and restore workflows
  - existing grouped families remain: `profile`, `alias`, `health`, `shell`

### Grouping rules

Use a grouped subcommand (default) when:
- the action belongs to a clear domain (`app`, `profile`, `recover`, etc.)
- the action is one of several related lifecycle verbs
- grouping improves discoverability in `gdf <domain> --help`

Use a top-level command only when:
- it is a high-frequency workflow entry point across domains
- grouping would make first-run or daily usage less discoverable

### Rename compatibility policy

- 1.x compatibility expectations are documented in [CLI Compatibility Policy](compatibility.md).

## Commands

### Initialization

#### `gdf init [repository]`

Initialize gdf or clone existing dotfiles repository.

`gdf init` also creates a placeholder `~/.gdf/generated/init.sh` so shell startup sourcing is safe before the first `gdf apply`.

```bash
gdf init                              # Create new repo
gdf init git@github.com:user/dots.git # Clone existing
```

#### `gdf init setup [flags]`

Run first-run setup with profile/app bootstrap.

| Flag | Description |
| ---- | ----------- |
| `-p, --profile <profile>` | Profile to bootstrap (default: `default`) |
| `--apps <csv>` | Comma-separated starter apps to create/add |
| `--json` | Output setup summary as JSON |

```bash
gdf init setup
gdf init setup --profile work --apps git,kubectl --json
```

---

### App Management

#### `gdf app add <app> [flags]`

Add an app bundle to a profile.

| Flag                      | Description                         |
| ------------------------- | ----------------------------------- |
| `-p, --profile <profile>` | Target profile (if omitted: auto-select one profile, or prompt when multiple) |
| `--from-recipe`           | Use built-in recipe                 |
| `--interactive`           | Show recipe suggestions and dependency prompts |
| `--apply`                 | Preview and apply the selected profile after adding (requires confirmation unless `--yes`) |

```bash
gdf app add kubectl -p sre
gdf app add backend-dev --from-recipe --interactive
gdf app add git -p base --apply
```

`gdf app add` updates desired configuration only by default. It does not mutate live targets unless `--apply` is set.

#### `gdf app remove <app> [flags]`

Remove an app bundle from a profile.

| Flag                      | Description                         |
| ------------------------- | ----------------------------------- |
| `-p, --profile <profile>` | Target profile (if omitted: auto-select one profile, or prompt when multiple) |
| `--uninstall`             | Unlink managed dotfiles and uninstall package only when no profiles still reference the app |
| `--dry-run`               | Preview removal actions without writing changes |
| `--yes`                   | Skip uninstall/unlink confirmation prompt |
| `--apply`                 | Preview and apply the selected profile after removal (requires confirmation unless `--yes`) |

When `--uninstall` is set, GDF prints a removal plan before applying changes.
When an app becomes unreferenced by all profiles, GDF prints cleanup guidance for `gdf app prune`.

#### `gdf app list [flags]`

List apps in a profile.

| Flag                      | Description                                    |
| ------------------------- | ---------------------------------------------- |
| `-p, --profile <profile>` | Profile to list apps from (if omitted: auto-select one profile, or prompt when multiple) |

#### `gdf app prune [flags]`

Archive or delete orphaned local app definitions (apps not referenced by any profile).

| Flag | Description |
| ---- | ----------- |
| `--dry-run` | Preview orphan cleanup actions without writing changes |
| `--delete` | Permanently delete orphaned app definitions and their dotfiles (default is archive) |
| `--json` | Output prune plan/result as JSON |
| `--yes` | Skip delete confirmation prompt in `--delete` mode |

```bash
gdf app prune --dry-run
gdf app prune
gdf app prune --delete --yes
```
 
#### `gdf app library`

Manage and explore the built-in app library.

##### `gdf app library list`

List all available recipes in the embedded library.

##### `gdf app library describe <recipe>`

Show the YAML definition and details of a specific recipe.

```bash
gdf app library describe git
```

#### `gdf app move <app-pattern> [flags]`
 
Move apps between profiles. At least one of `--from` or `--to` must be specified. If one side is omitted, GDF resolves it using existing profiles.
 
Supports wildcard patterns (e.g. `gnome-*`, `*`).
 
| Flag               | Description                         |
| ------------------ | ----------------------------------- |
| `--from <profile>` | Source profile (if omitted: auto-select one profile, or prompt when multiple) |
| `--to <profile>`   | Target profile (if omitted: auto-select one profile, or prompt when multiple) |
| `--apply`          | Preview and apply both affected profiles after move (requires confirmation unless `--yes`) |
 
```bash
gdf app move git --from work --to home
gdf app move "gnome-*" --to desktop
gdf app move --from old-work --to work # Move all apps
```

#### `gdf app install <app> [flags]`

Install an app directly. if the app is not defined or the installation method is unknown for the current OS, it will prompt to learn the package details.
If the app already defines `package.custom`, GDF uses that script as a valid install method when no package-manager mapping is available.

If `--profile` is omitted, GDF selects a profile automatically when exactly one exists, or prompts you when multiple profiles exist.

| Flag        | Description                                  |
| ----------- | -------------------------------------------- |
| `--package` | Specify package name manually (skips prompt) |
| `-p, --profile <profile>` | Profile to add app to (if omitted: auto-select one profile, or prompt when multiple) |

```bash
gdf app install ripgrep
gdf app install ripgrep --package ripgrep-cli
```

#### `gdf app track <path> [flags]`

Track existing dotfile and associate with an app.

| Flag              | Description                    |
| ----------------- | ------------------------------ |
| `-a, --app <app>` | App bundle to add this file to |
| `--secret`        | Mark file as secret (add to .gitignore) |
| `--interactive`   | Preview and resolve target/path conflicts interactively |

```bash
gdf app track ~/.kube/config -a kubectl
gdf app track ~/.gitconfig -a git
gdf app track ~/.aws/config -a aws-cli --secret
```

#### `gdf app import [paths...] [flags]`

Discover and adopt existing dotfiles, aliases, and common tool configs.

Import modes:
- `--preview`: preview-only discovery
- guided mapping (default in interactive terminals)
- `--apply`: apply import directly

| Flag | Description |
| ---- | ----------- |
| `--preview` | Preview discovered items without importing |
| `--apply` | Apply import directly |
| `--json` | Output preview/result as JSON |
| `-p, --profile <profile>` | Profile to add imported apps to (if omitted: auto-select one profile, or prompt when multiple) |
| `--sensitive-handling <ignore|secret|plain>` | Required in `--apply` mode when sensitive files are detected |

```bash
gdf app import --preview
gdf app import                    # guided mapping
gdf app import --apply --sensitive-handling secret
```

---

### Aliases

#### `gdf alias add <name> <command> [flags]`

Add a shell alias. If `--app` is specified, the alias is added to that app's bundle. Otherwise, GDF checks whether the command's first word matches an existing app bundle. If no match is found, the alias is stored as a global (unassociated) alias in `~/.gdf/aliases.yaml`.

| Flag              | Description                     |
| ----------------- | ------------------------------- |
| `-a, --app <app>` | App bundle to add this alias to |

```bash
gdf alias add k kubectl                 # auto-detects kubectl app
gdf alias add gco "git checkout" -a git  # explicit app
gdf alias add ll "ls -la"                # global (unassociated)
```

#### `gdf alias list`

List all aliases from all app bundles and global aliases. Aliases are grouped by app, with unassociated aliases shown separately.

#### `gdf alias remove <name>`

Remove an alias. Searches all app bundles and global aliases.

---

### Profiles

#### `gdf profile create <name> [flags]`

Create a new profile.

| Flag                   | Description         |
| ---------------------- | ------------------- |
| `--description <text>` | Profile description |

```bash
gdf profile create work
gdf profile create home --description "Home environment configuration"
```

#### `gdf profile list`

List all profiles with summary, showing app count and includes.

```bash
gdf profile list
```

#### `gdf profile show <name>`

Show detailed profile information including apps, includes, and conditions.

```bash
gdf profile show work
gdf profile show sre
```

#### `gdf profile delete <name>`

Delete a profile. The `default` profile cannot be deleted.

Delete mode flags (choose at most one):

| Flag | Description |
| ---- | ----------- |
| `--migrate-to-default` | Move apps from the deleted profile into `default` |
| `--purge` | Purge apps unique to the deleted profile (definition + managed cleanup) |
| `--leave-dangling` | Delete the profile and leave app definitions unreferenced |
| `--dry-run` | Preview profile deletion impact across apps, dotfiles, and packages without applying |
| `--yes` | Skip purge confirmation prompt |

If no mode flag is provided, behavior defaults to `--migrate-to-default`.
In interactive terminals, GDF prompts you to choose delete strategy when no mode flag is provided.
In `--non-interactive` mode, it remains deterministic and defaults to migrate.

```bash
gdf profile delete old-work
gdf profile delete old-work --purge --dry-run
gdf profile delete old-work --leave-dangling
```

#### `gdf profile rename <old-name> <new-name>`

Rename a profile. This command updates the profile directory, internal name, and updates any other profiles that include this profile to reference the new name.

```bash
gdf profile rename work work-2024
```

---

### Apply & Status

#### `gdf apply [profiles...]`

Apply one or more profiles to the system.

| Flag        | Description                                    |
| ----------- | ---------------------------------------------- |
| `--dry-run` | Show what would be done without making changes |
| `--allow-risky` | Proceed even if high-risk script patterns are detected |
| `--json` | Output dry-run plan as JSON (requires `--dry-run`) |
| `--run-apply-hooks` | Execute `hooks.apply` commands (disabled by default) |
| `--apply-hook-timeout <duration>` | Per-hook timeout when `--run-apply-hooks` is enabled (default: `30s`) |

This command performs the following operations:

1. **Resolve profile dependencies** - Processes profile `includes` in dependency order
2. **Resolve app dependencies** - Orders apps using topological sort
3. **Install packages** - Installs packages via package managers (when available)
4. **Link dotfiles** - Creates symlinks with conflict resolution
5. **Apply hooks (optional)** - Executes `hooks.apply` only when `--run-apply-hooks` is set; otherwise records deterministic skip details
6. **Generate shell integration** - Updates shell scripts for aliases/functions/env/init
7. **Generate managed completion files** - Writes app completion artifacts to `~/.gdf/generated/completions/{bash,zsh}/`
8. **Security scan** - Detects high-risk script patterns and requests confirmation before mutating operations
9. **Log operations** - Records all operations to `.operations/` for rollback
10. **Capture history snapshots** - Saves pre-change file snapshots to `.history/` before destructive replacements
11. **Update state** - Records applied profiles to `~/.gdf/state.yaml` (local only)

All operations are logged to `~/.gdf/.operations/<timestamp>.json`.
Historical snapshots are stored in `~/.gdf/.history/` and retained with quota-based eviction.
Non-dry-run apply acquires a run lock at `~/.gdf/.locks/apply.lock` to avoid concurrent apply corruption.

Package manager selection precedence for each app during apply:
1. `apps/*.yaml -> package.prefer`
2. `config.yaml -> package_manager.prefer`
3. platform auto-detection

When an app defines multiple package managers, GDF checks installed status across configured and available managers and skips reinstall when already satisfied.

If no profiles are provided, GDF resolves apply targets in this order:
1. Reuse profiles from local `~/.gdf/state.yaml` if present.
2. If exactly one profile exists, apply it automatically.
3. If multiple profiles exist, prompt you to choose one.
4. In `--non-interactive` mode with multiple profiles and no state, fail with guidance.

```bash
# Apply single profile
gdf apply base

# Apply multiple profiles (resolves includes)
gdf apply base programming sre

# Dry run to preview changes
gdf apply --dry-run work

# Profile with dependencies will include them
# If 'work' includes 'base', both are applied
gdf apply work

# Optional: add managed shell completion bootstrap from app library
gdf app add gdf-shell -p base
gdf apply base
```

#### `gdf status`

Show applied profiles, app summary, and drift overview.

| Flag | Description |
| ---- | ----------- |
| `--json` | Output status as JSON |

```bash
gdf status
gdf status --json
```

#### `gdf status diff`

Show detailed drift findings for managed targets.

| Flag | Description |
| ---- | ----------- |
| `--json` | Output drift details as JSON |
| `--patch` | Include unified patch output for non-symlink targets |
| `--max-bytes <n>` | Max source/target file size for patch generation (default: 1048576) |
| `--max-files <n>` | Max number of patches to generate per run (default: 20) |

```bash
gdf status diff
gdf status diff --patch
gdf status diff --patch --max-files 10 --max-bytes 262144
gdf status diff --json
```

**Output:**
```
Applied Profiles:
  ✓ base (5 apps) - applied 2 hours ago
  ✓ work (3 apps) - applied 30 minutes ago

Apps (8 total):
  git, zsh, vim, tmux, docker, kubectl, terraform, aws-cli

Last applied: 2024-02-11 18:30:00
```

**Behavior:**
- Shows applied profile timestamps and app counts
- Lists deduplicated app names
- Shows drift summary (source missing, target missing, mismatch, non-symlink)
- Suggests `gdf status diff` when drift exists
- If no profiles are applied, suggests using `gdf apply`

**State Tracking:**
- State is stored locally in `~/.gdf/state.yaml`
- State is LOCAL ONLY (gitignored) and does not sync across machines
- Updated automatically when `gdf apply` succeeds

---

### Recovery

#### `gdf recover rollback`

Undo the most recent operation log and restore captured historical snapshots when available.

| Flag | Description |
| ---- | ----------- |
| `--yes` | Skip confirmation prompt |
| `--choose-snapshot` | Prompt for a snapshot choice when multiple historical versions exist |
| `--target <path>` | Restore a specific target path from snapshot history |

```bash
# Rollback latest apply operations
gdf recover rollback

# Restore one file from historical snapshots
gdf recover rollback --target ~/.zshrc --choose-snapshot
```

#### `gdf recover restore [flags]`

Restore tracked files to their original locations and replace managed symlinks with real files.

| Flag | Description |
| ---- | ----------- |
| `--aliases-file <path>` | Path to export aliases to (default: `~/.aliases`) |

```bash
gdf recover restore
gdf recover restore --aliases-file ~/.aliases
```

---

### Git Operations

#### `gdf save [message]`

Stage all changes and commit.

```bash
# Save with default message "Update dotfiles"
gdf save

# Save with custom message
gdf save "Added kubectl config"
```

**Behavior:**
- Stages all changes in `~/.gdf` using `git add .`
- Creates a commit with the provided message or default "Update dotfiles"
- If no changes exist, displays informational message (not an error)
- Displays success message with commit summary

**Requirements:**
- Must be run after `gdf init`
- Git user must be configured (name and email)

#### `gdf push`

Push commits to the remote repository.

```bash
gdf push
```

**Behavior:**
- Pushes all commits to the configured remote
- Provides helpful error messages for common issues

**Requirements:**
- Remote must be configured: `git -C ~/.gdf remote add origin <url>`
- Upstream branch should be set (use `git -C ~/.gdf push -u origin main` first time)

**Common Errors:**
- "No git remote configured" - Add a remote first
- "No upstream branch" - Use `git -C ~/.gdf push -u origin main`

#### `gdf pull`

Pull changes from the remote repository.

```bash
gdf pull
```

**Behavior:**
- Pulls changes from the configured remote
- Merges changes into local repository

**Requirements:**
- Remote must be configured: `git -C ~/.gdf remote add origin <url>`

**Common Errors:**
- "No git remote configured" - Add a remote first
- "Merge conflict detected" - Resolve conflicts manually in `~/.gdf`

#### `gdf sync`

Full sync: pull, commit changes, and push.

```bash
gdf sync
```

**Behavior:**
1. Pulls changes from remote
2. Commits any local changes (if present)
3. Pushes to remote

This is the recommended workflow for keeping multiple machines in sync.

**Requirements:**
- Remote must be configured: `git -C ~/.gdf remote add origin <url>`
- Upstream branch should be set

**Example Workflow:**
```bash
# On machine 1: Make changes and sync
gdf app add kubectl
gdf app track ~/.kube/config
gdf sync

# On machine 2: Pull changes
gdf sync
gdf apply
```

---

### Maintenance

#### `gdf health validate`

Validate config, profile, and app YAML plus semantic integrity.

| Flag | Description |
| ---- | ----------- |
| `--json` | Output findings as JSON |

#### `gdf health doctor`

Run environment health checks (repo structure, shell integration, package manager availability, permissions).

| Flag | Description |
| ---- | ----------- |
| `--json` | Output findings as JSON |

#### `gdf health fix`

Apply safe, reviewable auto-fixes for common doctor findings.

| Flag | Description |
| ---- | ----------- |
| `--guarded` | Include higher-impact fixes that require backup-before-write behavior |
| `--dry-run` | Preview fix actions without applying changes |

#### `gdf health ci`

Run fail-fast validation and doctor checks for CI workflows.

| Flag | Description |
| ---- | ----------- |
| `--json` | Output combined findings as JSON |

#### `gdf shell reload`

Reload shell integration.

#### `gdf shell completion <bash|zsh>`

Generate shell completion script to stdout.

```bash
gdf shell completion bash > ~/.local/share/bash-completion/completions/gdf
gdf shell completion zsh > ~/.zfunc/_gdf
```

After generating, reload your shell or source the completion file according to your shell setup.

---

### Updates

#### `gdf update [flags]`

Check for updates and self-update GDF.

| Flag      | Description                                |
| --------- | ------------------------------------------ |
| `--never` | Disable auto-update checks permanently    |

```bash
gdf update          # Check and update to latest
gdf update --never  # Disable auto-checks
```
