# GDF - Go Dotfiles Manager

A cross-platform dotfile manager that unifies **packages, configuration files, and shell aliases** into coherent "app bundles."

## Features

- 🎯 **App Bundles** - Package + config + aliases managed together
- 🔗 **Composable Profiles** - Mix and match: base, work, programming, sre
- 🖥️ **Cross-Platform** - macOS, Linux, WSL with OS abstraction
- 📦 **Package Management** - Homebrew, apt, dnf with unified interface
- 🔄 **Git Backend** - Sync dotfiles across all your machines
- ⚡ **80/20 CLI** - Simple commands for common tasks, YAML for advanced

## Quick Start

```bash
# Install
go install github.com/rztaylor/GoDotFiles/cmd/gdf@latest

# Initialize with existing dotfiles repo
gdf init git@github.com:username/dotfiles.git

# Or start fresh
gdf init

# Apply profiles
gdf apply base programming

# Track existing config
gdf track ~/.gitconfig git

# Add alias (auto-associates with app)
gdf alias k kubectl

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

See [GEMINI.md](GEMINI.md) for AI-assisted development guidelines.

## License

Apache 2.0
