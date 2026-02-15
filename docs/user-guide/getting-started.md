# Getting Started with GDF

This guide walks you through setting up GDF on a new machine.

## Installation

The easiest way to install GDF on Linux or macOS is using the official install script:

```bash
curl -sfL https://raw.githubusercontent.com/rztaylor/GoDotFiles/main/scripts/install.sh | sh
```

### Alternative Methods

- **Manual**: Download the latest release from the [Releases page](https://github.com/rztaylor/GoDotFiles/releases).
- **From Source**: `go install github.com/rztaylor/GoDotFiles/cmd/gdf@latest` (requires Go 1.21+)

## Initial Setup

### Option 1: Clone Existing Dotfiles

If you already have a dotfiles repository:

```bash
gdf init git@github.com:username/dotfiles.git
```

This will:
1. Clone your repository to `~/.gdf/`
2. Detect available profiles
3. Prompt you to apply profiles

### Option 2: Start Fresh

```bash
gdf init
```

This creates a new empty repository at `~/.gdf/`.

## Shell Integration

Add this line to your `~/.bashrc` or `~/.zshrc`:

```bash
[ -f ~/.gdf/generated/init.sh ] && source ~/.gdf/generated/init.sh
```

Then reload your shell:

```bash
source ~/.zshrc  # or ~/.bashrc
```

## Your First Profile

Create a base profile:

```bash
gdf profile create base
```

Track some existing dotfiles:

```bash
gdf track ~/.gitconfig -a git
gdf track ~/.zshrc -a zsh
```

Add an alias:

```bash
gdf alias add ll "ls -la"
```

Apply your profile:

```bash
gdf apply base
```

If high-risk hook/script patterns are detected during apply, GDF will show the command content and ask for confirmation before proceeding.

If you need to undo the latest changes:

```bash
gdf rollback
```

## Next Steps

- [Learn about Profiles](profiles.md)
- [Understand App Bundles](apps.md)
- [Set up Git Syncing](syncing.md)
