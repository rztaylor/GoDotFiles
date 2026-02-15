# GDF Tutorial: Build a Portable, Version-Controlled Environment

This tutorial teaches both the **how** and the **why** of GDF (Go Dotfiles).

By the end, you will have:
- A reproducible environment defined in `~/.gdf`
- Cleanly organized app bundles for tools like Git, shell, and Kubernetes
- Profile-based setups you can apply differently on laptop, workstation, and server
- A Git-backed workflow to sync everything across machines

## Who This Is For

This guide is for engineers who work in terminals and care about setup quality over time:
- Developers who use multiple machines
- SRE/DevOps engineers who need reliable toolchains everywhere
- Anyone tired of rebuilding shell and editor setup by hand after reinstalls

If your stack differs from the examples, reuse the same pattern with your own apps.

## Why GDF Instead of Manual Dotfile Management?

| Without GDF                                   | With GDF                                              |
| --------------------------------------------- | ----------------------------------------------------- |
| Dotfiles are scattered under home directories | Dotfiles are centralized in `~/.gdf/dotfiles/`        |
| Setup steps are tribal knowledge              | Setup is encoded in app/profile YAML                  |
| Aliases/functions drift between machines      | Shell integration is generated from source of truth   |
| New machine setup is manual and error-prone   | `gdf init` + `gdf apply` restores environment quickly |
| Rollback is ad hoc                            | `gdf recover rollback` and history snapshots provide recovery |

GDF is not only a symlink tool. It models your environment as **app bundles + profiles + Git sync**, so your setup becomes a maintainable system.

## Core Concepts You Need First

### Dotfiles and symlinks

GDF tracks real config files inside `~/.gdf/dotfiles/<app>/...` and places symlinks at their expected locations.

Example:

```text
~/.gitconfig  ->  ~/.gdf/dotfiles/git/.gitconfig
```

Your tools still read `~/.gitconfig`, but the managed source lives in one place.

### App bundles

An app bundle describes everything about one tool:
- Package installation info
- Dotfiles to link
- Shell aliases/env/functions/init snippets
- Dependencies and optional hooks

Examples: `git`, `zsh`, `kubectl`, `terraform`.

### Profiles

A profile is a named collection of app bundles for a context.

Example profile strategy:
- `base`: essentials on every machine
- `programming`: languages and dev tooling
- `sre`: ops/cloud tooling
- `work`: includes `base` + `sre`

### Git backend

`~/.gdf` is a Git repo. You use `gdf save`, `gdf push`, `gdf pull`, and `gdf sync` to version and sync your environment.

## Quickstart: First Success in ~15 Minutes

### 1. Install GDF

Install script:

```bash
curl -sfL https://raw.githubusercontent.com/rztaylor/GoDotFiles/main/scripts/install.sh | sh
```

Or from source:

```bash
go install github.com/rztaylor/GoDotFiles/cmd/gdf@latest
```

Confirm:

```bash
gdf --help
```

### 2. Initialize your GDF repo

```bash
gdf init
```

Optional: initialize from an existing remote repo:

```bash
gdf init git@github.com:your-username/dotfiles.git
```

Optional guided bootstrap for first-run profile/apps:

```bash
gdf init setup --profile base --apps git,kubectl
```

### 3. Activate shell integration

`gdf init` already prompts to:
- Add the source line to your shell RC file.
- Enable event-based auto-reload on prompt (recommended, default: `Y`).
- Install shell completion for your detected shell (recommended, default: `Y`).

If you accept auto-reload, future `gdf apply` shell updates are picked up automatically on the next prompt in interactive bash/zsh sessions.

For your current terminal session:

```bash
source ~/.gdf/generated/init.sh
```

`gdf init` creates a placeholder `~/.gdf/generated/init.sh` so this command is safe immediately. Run `gdf apply` to generate the actual shell integration content.

If you skipped the prompt (or auto-injection failed), add this to your shell RC (`~/.zshrc` or `~/.bashrc`) and reload:

```bash
[ -f ~/.gdf/generated/init.sh ] && source ~/.gdf/generated/init.sh
source ~/.zshrc  # or ~/.bashrc
```

If you skipped completion setup during `gdf init`, you can install it manually:

```bash
# bash
gdf shell completion bash > ~/.local/share/bash-completion/completions/gdf

# zsh
mkdir -p ~/.zfunc
gdf shell completion zsh > ~/.zfunc/_gdf
```

For profile-managed completion bootstrap (recommended for reproducibility), add the library pseudo-app and apply:

```bash
gdf app add gdf-shell -p base
gdf apply base
```

### 4. Create a base profile

```bash
gdf profile create base --description "Essential tools for all machines"
```

### 5. Add one app and track one real dotfile

```bash
gdf app add git -p base
gdf app track ~/.gitconfig -a git
```

By default, `gdf app add` updates desired repo state only; it does not mutate your live system until apply. Use `--apply` for a guarded add-and-apply flow:

```bash
gdf app add git -p base --apply
```

If you omit `--profile` on profile-dependent commands, GDF now resolves it as follows:
- no profiles: returns an error
- one profile: selects it automatically
- multiple profiles: prompts you to choose

What this does:
1. Creates/updates the `git` app bundle.
2. Moves your managed copy into `~/.gdf/dotfiles/git/.gitconfig`.
3. Replaces `~/.gitconfig` with a symlink.

If you already have multiple files/aliases to adopt, run import first:

```bash
gdf app import --preview
gdf app import                  # guided mapping
gdf app import --apply --sensitive-handling secret
```

### 6. Add aliases through GDF

```bash
gdf alias add g git
gdf alias add gs "git status"
```

Because the command starts with `git` and `git` app exists, GDF associates these aliases with that app bundle.

### 7. Preview and apply

```bash
gdf apply --dry-run base
gdf apply base
```

### 8. Check state

```bash
gdf status
gdf status diff
gdf health doctor
gdf alias list
gdf profile show base
```

At this point, you already have the core GDF loop working.

## Why the New Health and Status Commands Matter

As your setup grows, two classes of problems become common:
- repository definition issues (invalid YAML, broken references, invalid conditions)
- machine drift (missing links, mismatched targets, missing generated shell integration)

The `status` and `health` command families exist to catch these issues early.

### `gdf status`

Use this as your high-level snapshot:
- what profiles are currently applied
- when they were last applied
- whether drift exists

### `gdf status diff`

Use this after `gdf status` reports drift. It explains exactly what changed, including:
- missing managed source files
- missing target files
- non-symlink targets
- symlink destination mismatches

For non-symlink drift, you can request patch-style output with bounded performance:

```bash
gdf status diff --patch
gdf status diff --patch --max-files 10 --max-bytes 262144
```

### `gdf health validate`

Use this before apply (and in CI) to validate repository definitions:
- config/profile/app schema and version checks
- semantic validation
- profile-to-app reference integrity

### `gdf health doctor`

Use this to verify machine readiness:
- repository structure/readability
- generated shell integration presence
- package manager detectability
- write permissions

### `gdf health fix`

Use this after `gdf health doctor` to apply safe, reviewable remediations for common setup issues.
Use `--guarded` to include higher-impact remediations that perform backup-before-write.

```bash
gdf health fix --guarded --dry-run
gdf health fix --guarded --yes
```

### `gdf health ci`

Use this in automation for fail-fast health checks:

```bash
gdf health ci --json --non-interactive
```

### Automation Flags

- `--json`: machine-readable output
- `--non-interactive`: never prompt; fail when confirmation is required
- `--yes`: auto-approve supported prompts

### Recommended Pre-Apply Safety Loop

```bash
gdf health validate
gdf health doctor
gdf apply --dry-run base
gdf status
```

## Build a Real Profile Strategy

Now extend from one-app proof to a full environment.

### Programming profile

```bash
gdf profile create programming --description "Languages and developer tooling"

gdf app add go -p programming
gdf app add python -p programming
gdf app add make -p programming
gdf app add ripgrep -p programming
gdf app add fzf -p programming

gdf app track ~/.config/pip/pip.conf -a python

gdf alias add gotest "go test ./..." -a go
gdf alias add py python3 -a python
gdf alias add pip "python3 -m pip" -a python
gdf alias add rg ripgrep -a ripgrep
```

### SRE profile

```bash
gdf profile create sre --description "SRE and cloud tooling"

gdf app add kubectl -p sre
gdf app add terraform -p sre
gdf app add docker -p sre
gdf app add aws-cli -p sre
gdf app add jq -p sre

gdf app track ~/.kube/config -a kubectl
gdf app track ~/.aws/config -a aws-cli --secret

gdf alias add k kubectl -a kubectl
gdf alias add kgp "kubectl get pods" -a kubectl
gdf alias add tf terraform -a terraform
```

> [!CAUTION]
> Use `--secret` for sensitive files so paths are tracked but content is gitignored. Never commit credentials.

### Compose profiles with includes

Create `work` that includes shared profiles:

```bash
gdf profile create work --description "Work machine"
```

Edit `~/.gdf/profiles/work/profile.yaml`:

```yaml
kind: Profile/v1
name: work
description: Work machine
includes:
  - base
  - sre
```

Now `gdf apply work` resolves and applies included profiles too.

## Apply Per Machine (Practical Use-Cases)

Examples:
- Personal laptop: `gdf apply base programming`
- Work laptop: `gdf apply base programming sre`
- Production bastion host: `gdf apply base sre`

Always preview first:

```bash
gdf apply --dry-run base sre
```

## Sync and Reproduce on Another Machine

### On your current machine

```bash
git -C ~/.gdf remote add origin git@github.com:your-username/dotfiles.git
gdf save "Bootstrap profiles and core dotfiles"
git -C ~/.gdf push -u origin main
```

After first push, daily sync is usually:

```bash
gdf sync
```

### On a second machine

```bash
# Install gdf first, then:
gdf init git@github.com:your-username/dotfiles.git
gdf apply --dry-run base sre
gdf apply base sre
source ~/.gdf/generated/init.sh
```

For completion on the second machine, run `gdf shell completion bash` or `gdf shell completion zsh` once during setup.

This is the core payoff: your setup is reproducible, not rebuilt manually.

## Day-to-Day Workflow

Typical loop:

```bash
# Add or change config
gdf app track ~/.tmux.conf -a tmux
gdf alias add t "tmux attach || tmux"

# Apply locally if needed
gdf apply base

# Commit and sync
gdf save "Add tmux config and alias"
gdf push
# or: gdf sync
```

On another machine:

```bash
gdf sync
gdf apply base
```

## Safety and Recovery

Use these consistently:
- `gdf apply --dry-run ...` before every apply
- `gdf health validate` to catch definition issues before changes
- `gdf health doctor` to verify machine readiness
- `gdf status` to verify applied state and drift summary
- `gdf status diff` to inspect exact drift details
- `gdf recover rollback` to undo last operation log
- `gdf recover rollback --target ~/.zshrc --choose-snapshot` to restore one file from history
- `gdf recover restore` if you need to replace managed symlinks with real files at original paths

## Advanced: Edit App Bundles Directly

CLI commands cover common workflows; YAML gives full control.

### Example app bundle

`~/.gdf/apps/kubectl.yaml`:

```yaml
kind: App/v1
name: kubectl
description: Kubernetes command-line tool

package:
  brew: kubectl
  apt: kubectl
  dnf: kubernetes-client

dotfiles:
  - source: kube/config
    target: ~/.kube/config

shell:
  aliases:
    k: kubectl
    kgp: kubectl get pods
  env:
    KUBECONFIG: ~/.kube/config
```

### Example dependency-only meta bundle

`~/.gdf/apps/backend-dev.yaml`:

```yaml
kind: App/v1
name: backend-dev
description: Grouped backend tooling

dependencies:
  - go
  - docker
  - kubectl
  - terraform
```

Then add it to a profile:

```bash
gdf app add backend-dev -p programming
```

GDF resolves dependencies during apply, so this works as a reusable grouped capability.

## Optional: Direct Install Shortcut

When you want to install a tool immediately:

```bash
gdf app install ripgrep -p programming
```

If package metadata is missing for your platform, GDF can prompt to learn installation details and persist them in the app definition.

## Common Mistakes to Avoid

- Tracking secret material without `--secret`
- Applying without `--dry-run`
- Putting every app in one profile instead of context-based profiles
- Disabling auto-reload and then forgetting to manually reload shell after alias/init changes
- Using vague commit messages for environment changes

## Next Steps

- Explore built-in recipes: `gdf app library list` and `gdf app library describe <recipe>`
- Troubleshoot common setup, sync, recovery, and offboarding issues: [FAQ](FAQ.md)
- Learn full command behavior: [CLI Reference](../reference/cli.md)
- Learn all YAML fields and conditions: [YAML Schema Reference](../reference/yaml-schemas.md)
- Review design rationale: [Architecture Overview](../architecture/overview.md)
