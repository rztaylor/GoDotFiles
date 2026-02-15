# Getting Started with GDF

This page is the fastest path to a working GDF setup.

If you want deeper concepts, profile design strategy, and multi-machine workflows, continue to the [full tutorial](tutorial.md).

## 1. Install

Install script (Linux/macOS):

```bash
curl -sfL https://raw.githubusercontent.com/rztaylor/GoDotFiles/main/scripts/install.sh | sh
```

Alternative methods:
- Releases: [GitHub Releases](https://github.com/rztaylor/GoDotFiles/releases)
- Source install (Go 1.21+): `go install github.com/rztaylor/GoDotFiles/cmd/gdf@latest`

Verify:

```bash
gdf --help
```

## 2. Initialize Repository

Start fresh:

```bash
gdf init
```

Or clone existing dotfiles:

```bash
gdf init git@github.com:username/dotfiles.git
```

GDF stores everything under `~/.gdf/`.

## 3. Activate Shell Integration

`gdf init` prompts to:
- Add the shell source line to your RC file.
- Enable event-based auto-reload on prompt (recommended).
- Install shell completion for your detected shell (recommended).

For the current terminal session, just load generated integration:

```bash
source ~/.gdf/generated/init.sh
```

If you skipped the init prompt (or auto-injection failed), add this to your shell RC (`~/.zshrc` or `~/.bashrc`) and reload:

```bash
[ -f ~/.gdf/generated/init.sh ] && source ~/.gdf/generated/init.sh
source ~/.zshrc  # or ~/.bashrc
```

## 4. Create a Base Profile

```bash
gdf profile create base --description "Essential tools for every machine"
```

## 5. Add First App + Dotfile

Create the app bundle, then track your config:

```bash
gdf app add git -p base
gdf app track ~/.gitconfig -a git
```

This moves your managed copy into `~/.gdf/dotfiles/git/.gitconfig` and symlinks `~/.gitconfig` to it.

## 6. Add a Couple of Aliases

```bash
gdf alias add g git
gdf alias add gs "git status"
```

## 7. Preview, Apply, Verify

Preview changes first:

```bash
gdf apply --dry-run base
```

Apply:

```bash
gdf apply base
```

Verify:

```bash
gdf status
gdf status diff
gdf health doctor
gdf profile show base
gdf alias list
```

## 8. Save and Sync (Optional but Recommended)

```bash
gdf save "Initial GDF setup"
```

If remote is configured:

```bash
gdf push
# or use: gdf sync
```

## Recovery Commands

If something goes wrong:

```bash
gdf recover rollback
```

To restore managed files back to regular files at original locations:

```bash
gdf recover restore
```

To validate configuration and run a safe auto-repair pass:

```bash
gdf health validate
gdf health fix
```

To include guarded repair actions (with backups), preview first:

```bash
gdf health fix --guarded --dry-run
```

To clean up orphaned app definitions (apps no longer referenced by any profile):

```bash
gdf app prune --dry-run
gdf app prune
```

## Next Steps

- Full walkthrough with concepts and use-cases: [Tutorial](tutorial.md)
- All commands and flags: [CLI Reference](../reference/cli.md)
- YAML fields and schema versions: [YAML Schema Reference](../reference/yaml-schemas.md)
- Architecture and design context: [Architecture Overview](../architecture/overview.md)
