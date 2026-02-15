# Changelog

All notable changes to GDF will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Add `gdf app add --apply` for a safer add-and-apply flow with preview and confirmation.
- Add a new user FAQ with practical setup, conflict, sync, recovery, and offboarding scenarios.

### Changed
- `gdf apply` with no profile arguments now reuses previously applied profiles when available, then falls back to profile selection.
- Clarify that `gdf app add` updates configuration only; run `gdf apply` to change the live system.

### Fixed
- Profile-based commands no longer assume `default` when `--profile` is omitted: they now error with no profiles, auto-select with one, and prompt with many.
- `gdf app add` is more reliable when starting from a minimal or partially initialized repository.
- `gdf apply` package installation is smarter about manager preferences and avoids reinstalling tools already present through another supported manager.
- Fresh and repaired config now expose all major settings by default, making options easier to discover and edit.
- Init and cloned-repo setup now restore expected baseline folders/files more consistently, reducing immediate post-init health issues.
- `gdf apply` no longer emits spurious completion warnings for `git`.
- Base recipes now provide better out-of-the-box coverage for common tools (including `git`, `zsh`, `just`, and `oh-my-zsh`) with safer first-time apply behavior.
- Apps that use only custom install scripts are now treated as valid installable definitions instead of being rejected as unsupported.

## [1.1.0] - 2026-02-15

### Added
- Add `gdf-shell` library pseudo-app to bootstrap profile-managed shell completion sourcing from `~/.gdf/generated/completions/{bash,zsh}`.

### Changed
- Expand library completion support by adding completion generation for `git`, `kubectl`, `gh`, `helm`, and `docker` recipes.
- Generate app `shell.completions` outputs during `gdf apply` into managed completion files instead of relying on runtime completion command sourcing in `init.sh`.

### Fixed
- Create placeholder `~/.gdf/generated/init.sh` during `gdf init` so `source ~/.gdf/generated/init.sh` does not fail before first `gdf apply`.

## [1.0.0] - 2026-02-15

### Added
- Add `gdf init setup` for guided first-run profile/app bootstrap with optional JSON summary output.
- Add `gdf app import` with preview, guided mapping, and apply modes to adopt existing dotfiles and aliases.
- Add conflict decision audit logs under `.operations/decisions-*.json` for interactive import/track conflict handling.
- Add `gdf shell completion <bash|zsh>` to generate shell completion scripts for interactive shell setup.
- Add optional event-based shell auto-reload hooks so updated shell integration is picked up automatically on the next prompt.
- Add `gdf health validate`, `gdf health doctor`, `gdf health fix`, and `gdf health ci` for validation, diagnostics, safe repair, and CI health checks.
- Add `gdf status diff` to show detailed drift findings for managed targets.
- Add stable CLI exit codes for health and non-interactive failure paths to make scripting more reliable.
- Add `gdf app remove --uninstall` cleanup flow with preview and guarded confirmation, including package uninstall only when uniquely owned by the removed app.
- Add `gdf profile delete` mode flags (`--migrate-to-default`, `--purge`, `--leave-dangling`) with `--dry-run` impact previews.
- Add `gdf app prune` to detect orphaned app definitions and archive them by default, with optional permanent delete mode.
- Add optional patch-style output to `gdf status diff` with explicit `--max-files` and `--max-bytes` limits.
- Add `gdf health fix --guarded` and `gdf health fix --dry-run` for previewable higher-impact remediations with backup-before-write behavior.
- Add `gdf-shell` library pseudo-app and apply-time managed completion artifacts under `~/.gdf/generated/completions/{bash,zsh}` for profile-managed shell completion bootstrap.

### Changed
- Add `--interactive` mode to `gdf app add` for recipe suggestions and dependency prompts.
- Add `--interactive` mode to `gdf app track` with explicit conflict previews and resolution choices.
- Reorganize app lifecycle commands under `gdf app` (`add`, `remove`, `list`, `install`, `track`, `move`, `library`) and recovery commands under `gdf recover` (`rollback`, `restore`), while keeping `init`, `save`, `push`, `pull`, and `sync` as top-level commands.
- Improve the getting started guide with a clearer quickstart flow and valid follow-up documentation links.
- Update `gdf init` shell onboarding to ask whether to enable event-based auto-reload (default yes) for faster out-of-the-box shell updates.
- Update `gdf init` onboarding to offer shell completion installation for the detected shell by default, with manual commands as fallback.
- Add `--json` output for `gdf status`, `gdf status diff`, `gdf health validate`, `gdf health doctor`, and `gdf health ci`.
- Add global `--yes` and `--non-interactive` flags for deterministic prompt handling in health and risky-apply flows.
- Add `gdf apply --dry-run --json` plan output for automation workflows.
- Expand prompt handling coverage so more interactive command paths follow deterministic `--non-interactive` and `--yes` behavior.
- Add interactive strategy selection for `gdf profile delete` when no delete mode flag is provided.
- Add orphan cleanup guidance after `gdf app remove` when an app is no longer referenced by any profile.
- Improve `gdf status diff` responsiveness for repeated runs by caching drift preview metadata.
- Update app `shell.completions` handling to generate managed completion files during `gdf apply` instead of embedding runtime completion commands in generated init scripts.
- Expand library completion support by adding managed shell completion commands for `gh`, `helm`, and `docker` recipes.

### Fixed
- Require `gdf health` commands to run only inside an initialized `~/.gdf` repository, matching other repo-dependent commands.
- Ensure `gdf status` reports an initialization error when `~/.gdf` is not initialized.
- Create a placeholder `~/.gdf/generated/init.sh` during `gdf init` so immediate `source ~/.gdf/generated/init.sh` no longer fails before first `gdf apply`.

## [0.8.0] - 2026-02-15

### Added
- **Safer Apply for Risky Scripts**: `gdf apply` now warns and asks for confirmation when app hooks or install scripts look high risk (for example remote script pipe-to-shell patterns).
- **Rollback Command**: Added `gdf recover rollback` so you can undo the latest apply changes and recover files from captured history.
- **Targeted File Recovery**: `gdf recover rollback --target <path> --choose-snapshot` lets you restore a specific file and choose from dated historical versions.
- **More Flexible Conditional Dotfiles**: Dotfile conditions now support boolean expressions (`AND`/`OR`) and parentheses.
- **Better Cross-Platform Dotfile Targets**: Dotfiles can now cleanly map different target paths per OS from a single app definition.

### Changed
- **Improved Shell Startup Setup**: Generated shell integration now includes startup initialization for supported tools, reducing manual shell config edits.
- **Expanded Base Recipe Experience**: Common tools like `direnv`, `starship`, `zoxide`, and `fzf` now work with less manual shell setup.
- **Stronger File Safety During Apply**: GDF now keeps historical file copies before destructive replacements, improving rollback reliability.
- **Snapshot Retention Control**: Added `history.max_size_mb` (default `512`) so users can tune how much disk space history snapshots can use.

### Fixed
## [0.7.0] - 2026-02-14

### Added

- **App Library**: Expanded core library with 30+ new app recipes (Languages, DevOps, Utilities) including "Base Profile" tools like git, zsh, starship.
- **Library CLI**: New `gdf app library list` and `describe` commands to explore embedded recipes.
- **Interactive Add**: `gdf app add` now prompts to use recipes if available, or to confirm skeleton creation for new apps.
- **No-Override Safety**: `gdf app add` explicitly warns and preserves existing app configurations instead of overriding them.
- **Meta-Profiles**: Support for grouping tools via meta-apps (e.g., `backend-dev`).
- **Recursive Dependencies**: `gdf apply` now recursively installs dependencies for apps in profiles.
- **Self-Update**: Implemented `gdf update` command and periodic auto-update checks.
- **Install Script**: Added `scripts/install.sh` for easy installation.

### Changed

### Fixed

## [0.6.0] - 2026-02-14


### Added

- **Schema Versioning**: Enforced `kind` field (e.g., `App/v1`) in all YAML files to support future schema evolution.
- **Restore Command**: Added `gdf recover restore` to revert changes, restore files, and export aliases for safe uninstallation.
- **Version Command**: Added `gdf version` to output build information.
- **License**: Switched to Apache 2.0 License.
- **Docs**: Cleaned up broken links in documentation.
- **Safety**: Enforce initialization check for all commands (except `version`, `init`, and `help`) to ensure GDF is run within a valid repository.
- **Release Automation**: Added `make release` command and GitHub Action configuration to automate releases via GoReleaser.
- **Documentation**: Added `CONTRIBUTING.md` and `docs/development/release.md` to guide contributors and maintainers.


## [0.5.0] - 2026-02-13

### Added

- **Core Features**
  - `gdf init` to initialize or clone dotfiles repositories (supports existing repos).
  - `gdf app add <app>` to add app bundles to profiles.
  - `gdf app track <path>` to adopt existing dotfiles into the repo.
  - `gdf apply` to link dotfiles and install packages from profiles.
  - `gdf app list`, `gdf app remove`, and `gdf show` for managing apps and profiles.
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
  - **Secret Tracking**: `gdf app track --secret` to safely add sensitive files to `.gitignore`.
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
1. Only include user-facing changes rather than internal changes.
2. Add entry under `[Unreleased]` in appropriate category
3. Use imperative mood ("Add" not "Added")
4. Reference issue/PR numbers if applicable
5. Keep entries concise but descriptive
