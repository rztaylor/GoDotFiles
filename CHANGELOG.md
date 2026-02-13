# Changelog

All notable changes to GDF will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2026-02-13

### Added

- **Core Features**
  - `gdf init` to initialize or clone dotfiles repositories (supports existing repos).
  - `gdf add <app>` to add app bundles to profiles.
  - `gdf track <path>` to adopt existing dotfiles into the repo.
  - `gdf apply` to link dotfiles and install packages from profiles.
  - `gdf list`, `gdf remove`, and `gdf show` for managing apps and profiles.
  - **Auto-detection** of app names when tracking files (e.g. `.gitconfig` -> `git`).

- **Profile Management**
  - **Profiles**: Support for multiple profiles (e.g., `default`, `work`, `home`).
  - **Conditions**: Conditional profile inclusion based on OS, Distro, Hostname, and Arch.
  - **Smart Defaults**: Commands default to `default` profile if not specified.
  - **Dependencies**: Profiles can include other profiles.

- **Shell & Terminal**
  - **Aliases**: `gdf alias` management with global and profile-specific aliases.
  - **Shell Integration**: Auto-generating shell configuration for bash and zsh.
  - **Global Aliases**: Support for aliases not tied to specific apps (e.g. `ls`, `cd`).

- **Package Management**
  - **Multi-Platform**: Support for Homebrew/Linuxbrew, APT, and DNF.
  - **Custom Scripts**: Execute custom install/uninstall scripts with security prompts.

- **Git & State**
  - **Git Operations**: `gdf save`, `gdf push`, `gdf pull`, `gdf sync` for easy repo management.
  - **Status**: `gdf status` to see what is applied and when.

- **Safety & Security**
  - **Secret Tracking**: `gdf track --secret` to safely add sensitive files to `.gitignore`.
  - **Backups**: Automatic backups of existing files during `apply`.
  - **Conflict Resolution**: Strategies for handling file conflicts (`backup_and_replace`, `error`).

### Fixed

- **Cross-Platform**: Improved package resolution logic for Linux vs macOS.


---

## Release Guidelines

### Version Numbering

- **MAJOR** (1.0.0): Breaking changes to CLI or config format
- **MINOR** (0.1.0): New features, backward compatible
- **PATCH** (0.1.1): Bug fixes, no new features

### Changelog Categories

- **Added**: New features
- **Changed**: Changes in existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Vulnerability fixes

### For AI Agents

When completing a task:
1. Add entry under `[Unreleased]` in appropriate category
2. Use imperative mood ("Add" not "Added")
3. Reference issue/PR numbers if applicable
4. Keep entries concise but descriptive
