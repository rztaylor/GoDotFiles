# Getting Started with GDF

This guide walks you through setting up GDF on a new machine.

## Installation

### From Source

```bash
go install github.com/rztaylor/GoDotFiles/cmd/gdf@latest
```

### From Binary (coming soon)

```bash
# macOS
brew install gdf

# Linux
curl -sSL https://gdf.dev/install.sh | bash
```

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
gdf track ~/.gitconfig git
gdf track ~/.zshrc zsh
```

Add an alias:

```bash
gdf alias ll "ls -la"
```

## Next Steps

- [Learn about Profiles](profiles.md)
- [Understand App Bundles](apps.md)
- [Set up Git Syncing](syncing.md)
