# GDF - Go Dotfiles Manager

A cross-platform dotfile manager that unifies **packages, configuration files, and shell integration** into coherent "app bundles."

## Features

- 🎯 **App Bundles** - Package + config + shell integration managed together
- 🔗 **Composable Profiles** - Mix and match: base, work, programming, sre
- 🖥️ **Cross-Platform** - macOS, Linux, WSL with OS abstraction
- 📦 **Package Management** - Homebrew, apt, dnf with unified interface
- 🔄 **Git Backend** - Sync dotfiles across all your machines
- 🛡️ **Security-Aware Apply** - Warns and confirms when high-risk script patterns are detected
- 🕘 **Operation History & Rollback** - Operation logs plus historical snapshots for safer recovery
- ⚡ **80/20 CLI** - Simple commands for common tasks, YAML for advanced

## Installation

The easiest way to install GDF is using the official install script:

```bash
curl -sfL https://raw.githubusercontent.com/rztaylor/GoDotFiles/main/scripts/install.sh | sh
```

### Alternative Methods

- **Manual**: Download binaries from the [Releases page](https://github.com/rztaylor/GoDotFiles/releases).
- **From Source**: `go install github.com/rztaylor/GoDotFiles/cmd/gdf@latest`

## Auto-Update

GDF automatically checks for updates every 24 hours.
- **Update now**: `gdf update`
- **Disable checks**: `gdf update --never`

## Quick Start

```bash
# 1. Install GDF
curl -sfL https://raw.githubusercontent.com/rztaylor/GoDotFiles/main/scripts/install.sh | sh

# Initialize with existing dotfiles repo
gdf init git@github.com:username/dotfiles.git

# Or start fresh
gdf init

# Apply profiles
gdf apply base programming

# Track existing config
gdf app track ~/.gitconfig -a git

# Add alias (auto-associates with app)
gdf alias add k kubectl

# Roll back latest apply if needed
gdf recover rollback

# Save and sync
gdf save "Added kubectl alias"
gdf push
```

## Documentation

- [Getting Started](docs/user-guide/getting-started.md)
- [Tutorial](docs/user-guide/tutorial.md) - Hands-on guide from zero to synced dotfiles
- [User Guide](docs/user-guide/)
- [CLI Reference](docs/reference/cli.md)
- [Architecture](docs/architecture/overview.md)

## Development

- See [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to this project.
- See [GEMINI.md](GEMINI.md) for AI-assisted development guidelines.

## License

Apache 2.0
