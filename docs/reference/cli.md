# CLI Reference

Complete reference for all gdf commands.

> [!TIP]
> All CLI output and interactive components should follow the [UI Style Guide](ui-style-guide.md).

## Global Flags

| Flag            | Description               |
| --------------- | ------------------------- |
| `-v, --verbose` | Enable verbose output     |
| `-h, --help`    | Show help for any command |

## Commands

### Initialization

#### `gdf init [repository]`

Initialize gdf or clone existing dotfiles repository.

```bash
gdf init                              # Create new repo
gdf init git@github.com:user/dots.git # Clone existing
```

---

### App Management

#### `gdf add <app> [flags]`

Add an app bundle to a profile.

| Flag                      | Description                         |
| ------------------------- | ----------------------------------- |
| `-p, --profile <profile>` | Target profile (default: `default`) |
| `--with <companions>`     | Include companion apps              |
| `--track <path>`          | Track config file                   |
| `--aliases`               | Include suggested aliases           |
| `--from-library`          | Use built-in recipe                 |

```bash
gdf add kubectl -p sre
gdf add kubectl --with kubectx,kubens --aliases
```

#### `gdf remove <app> [flags]`

Remove an app bundle from a profile.

| Flag                      | Description                         |
| ------------------------- | ----------------------------------- |
| `-p, --profile <profile>` | Target profile (default: `default`) |

#### `gdf list [flags]`

List apps in a profile.

| Flag                      | Description                                    |
| ------------------------- | ---------------------------------------------- |
| `-p, --profile <profile>` | Profile to list apps from (default: `default`) |
 
#### `gdf library`

Manage and explore the built-in app library.

##### `gdf library list`

List all available recipes in the embedded library.

##### `gdf library describe <recipe>`

Show the YAML definition and details of a specific recipe.

```bash
gdf library describe git
```

#### `gdf move <app-pattern> [flags]`
 
Move apps between profiles. At least one of `--from` or `--to` must be specified. If one is omitted, it defaults to `default`.
 
Supports wildcard patterns (e.g. `gnome-*`, `*`).
 
| Flag               | Description                         |
| ------------------ | ----------------------------------- |
| `--from <profile>` | Source profile (default: `default`) |
| `--to <profile>`   | Target profile (default: `default`) |
 
```bash
gdf move git -p work --to home
gdf move "gnome-*" --to desktop
gdf move --from old-work --to work # Move all apps
```

#### `gdf install <app> [flags]`

Install an app directly. if the app is not defined or the installation method is unknown for the current OS, it will prompt to learn the package details.

If the app is not part of any profile, it will be added to the `default` profile (unless `--profile` is specified).

| Flag        | Description                                  |
| ----------- | -------------------------------------------- |
| `--package` | Specify package name manually (skips prompt) |
| `-p, --profile <profile>` | Profile to add app to (default: `default`) |

```bash
gdf install ripgrep
gdf install ripgrep --package ripgrep-cli
```

#### `gdf track <path> [flags]`

Track existing dotfile and associate with an app.

| Flag              | Description                    |
| ----------------- | ------------------------------ |
| `-a, --app <app>` | App bundle to add this file to |
| `--secret`        | Mark file as secret (add to .gitignore) |

```bash
gdf track ~/.kube/config -a kubectl
gdf track ~/.gitconfig -a git
gdf track ~/.aws/config -a aws-cli --secret
```

---

### Aliases & Functions

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

#### `gdf fn <name>`

Add a shell function. Opens editor with template.

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

Delete a profile. If the profile contains apps, they are automatically moved to the `default` profile. The `default` profile cannot be deleted.

```bash
gdf profile delete old-work
```

#### `gdf profile rename <old-name> <new-name>`

Rename a profile. This command updates the profile directory, internal name, and updates any other profiles that include this profile to reference the new name.

```bash
gdf profile rename work work-2024
```

---

### Apply & Sync

#### `gdf apply <profiles...>`

Apply one or more profiles to the system.

| Flag        | Description                                    |
| ----------- | ---------------------------------------------- |
| `--dry-run` | Show what would be done without making changes |

This command performs the following operations:

1. **Resolve profile dependencies** - Processes profile `includes` in dependency order
2. **Resolve app dependencies** - Orders apps using topological sort
3. **Install packages** - Installs packages via package managers (when available)
4. **Link dotfiles** - Creates symlinks with conflict resolution
5. **Run apply hooks** - Executes hooks for package-less bundles
6. **Generate shell integration** - Updates shell scripts for aliases/functions/env/init/completions
7. **Log operations** - Records all operations to `.operations/` for rollback
8. **Update state** - Records applied profiles to `~/.gdf/state.yaml` (local only)

All operations are logged to `~/.gdf/.operations/<timestamp>.json`.

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
```

#### `gdf status`

Show which profiles are currently applied and their status.

```bash
gdf status
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
- Shows all applied profiles with app counts and timestamps
- Lists all apps across all applied profiles (deduplicated)
- Displays when profiles were last applied
- If no profiles are applied, suggests using `gdf apply`

**State Tracking:**
- State is stored locally in `~/.gdf/state.yaml`
- State is LOCAL ONLY (gitignored) and does not sync across machines
- Updated automatically when `gdf apply` succeeds


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
gdf add kubectl
gdf track ~/.kube/config
gdf sync

# On machine 2: Pull changes
gdf sync
gdf apply
```

---

### Maintenance

#### `gdf doctor`

Check system health and report issues.

#### `gdf shell reload`

Reload shell integration.

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
