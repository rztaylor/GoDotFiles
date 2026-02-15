# GDF Tutorial: From Zero to Fully Managed Dotfiles

This hands-on tutorial walks you through using **gdf** (Go Dotfiles) to manage your development environment end-to-end. By the end, you'll have a portable, version-controlled setup that can be deployed to any new machine in minutes.

> **Who is this for?** We'll use a realistic scenario throughout: you're a DevOps/SRE engineer who also writes Python and Go, working across a personal laptop and a work server. If your tooling differs, the concepts transfer directly — just swap the app names.

## What Are Dotfiles?

Most command-line tools store their configuration in plain text files in your home directory. These files usually start with a dot (`.`), which makes them hidden by default — hence the name "dotfiles." You've almost certainly got some already:

| File                      | What it configures                             |
| ------------------------- | ---------------------------------------------- |
| `~/.gitconfig`            | Git author name, email, aliases, diff settings |
| `~/.zshrc` or `~/.bashrc` | Shell prompt, aliases, `PATH`, startup scripts |
| `~/.vimrc`                | Vim editor settings, key bindings, plugins     |
| `~/.tmux.conf`            | Tmux prefix key, pane behaviour, status bar    |
| `~/.kube/config`          | Kubernetes cluster connection details          |
| `~/.aws/config`           | AWS CLI profiles, default region               |

Over time, you invest hours tweaking these files. They become a very personal part of your workflow — and losing them (or having to recreate them on a new machine) is painful.

### The problem

Dotfiles are scattered all over your home directory. Each tool expects its config in a specific location (`~/.gitconfig`, `~/.kube/config`, etc.), and there's no standard way to:

- **Back them up** — they're just loose files, easily lost if a disk fails or you reinstall your OS.
- **Sync them** — if you work on multiple machines (a laptop and a server, for instance), keeping configs consistent requires manual copying.
- **Version-control them** — you can't easily see what you changed, when, or roll back a mistake.

This is what "dotfile management" solves: gather all your dotfiles into a single, version-controlled repository, while keeping them in the locations each tool expects.

### How GDF manages dotfiles: tracking and symlinks

GDF stores **all** your managed dotfiles together inside `~/.gdf/dotfiles/`. But your tools still need to find their configs in the usual places (`~/.gitconfig`, `~/.vimrc`, etc.). GDF bridges this gap using **symlinks** — a standard filesystem feature where a file at one path transparently points to a file at another path.

Here's what happens when you run `gdf track ~/.gitconfig -a git`:

1. **Copy**: GDF copies `~/.gitconfig` into `~/.gdf/dotfiles/git/.gitconfig`. This is the managed copy, safely inside the GDF repository.
2. **Replace with symlink**: GDF replaces the original `~/.gitconfig` with a symlink that points to `~/.gdf/dotfiles/git/.gitconfig`.

The result looks like this:

```
~/.gitconfig  →  (symlink)  →  ~/.gdf/dotfiles/git/.gitconfig
~/.vimrc      →  (symlink)  →  ~/.gdf/dotfiles/vim/.vimrc
~/.tmux.conf  →  (symlink)  →  ~/.gdf/dotfiles/tmux/.tmux.conf
```

From Git's perspective, all your dotfiles live neatly inside `~/.gdf/dotfiles/`, ready to be committed and pushed. From your tools' perspective, nothing has changed — `~/.gitconfig` still exists exactly where Git expects it. The symlink is completely transparent.

> [!NOTE]
> If you've ever used tools like GNU Stow, chezmoi, or yadm, this is a similar concept. GDF's twist is that it groups dotfiles **by app** rather than treating them as a flat collection of files.

### App bundles: keeping related things together

In practice, a tool is more than just its config file. When you set up `git` on a new machine, you're really doing three things: installing the package, placing your config files, and setting up your favourite aliases. These belong together.

An **app bundle** is GDF's way of grouping everything related to a single tool into one unit:

| Component    | What it does                                               | Example (git)                         |
| ------------ | ---------------------------------------------------------- | ------------------------------------- |
| **Package**  | The software itself, installed via your OS package manager | `git` via brew, apt, or dnf           |
| **Dotfiles** | Configuration files, tracked and symlinked                 | `~/.gitconfig`, `~/.gitignore_global` |
| **Aliases**  | Shell shortcuts you use daily                              | `g` → `git`, `gs` → `git status`      |

When you run `gdf add git`, GDF creates a bundle definition file at `~/.gdf/apps/git.yaml`. This file is the single source of truth for everything about Git in your environment. You then build it up by tracking config files (`gdf track`) and adding aliases (`gdf alias`).

Here's what a fleshed-out bundle looks like on disk:

```
~/.gdf/
├── apps/
│   └── git.yaml            ← bundle definition (package, aliases, dotfile list)
├── dotfiles/
│   └── git/
│       ├── .gitconfig       ← your actual gitconfig content
│       └── .gitignore_global
└── ...
```

The power of this approach is portability. When you apply this bundle on a new machine, GDF can install the package, symlink the dotfiles into place, and generate your aliases — all from that single `git.yaml` definition.

### Profiles: composing your environment

You probably don't want the exact same tools on every machine. Your work laptop needs Kubernetes and Terraform, but your personal machine doesn't. Your dev server needs Go and Python, but maybe not a terminal multiplexer.

A **profile** is a named collection of app bundles that represent a particular context or role:

```
base          →  git, zsh, vim, tmux, curl         (every machine)
programming   →  go, python, make, ripgrep, fzf    (dev machines)
sre           →  kubectl, terraform, docker, aws-cli (work machines)
```

You choose which profiles to apply on each machine. On your work laptop you might apply all three; on a CI server, just `base`.

Profiles can also **include** other profiles. For example, you can create a `work` profile that includes `base` and `sre`, so running `gdf apply work` automatically pulls in everything from both.

### Git backend: version control and sync

Your entire `~/.gdf/` directory is a Git repository. Every change — a new app bundle, a tracked dotfile, an added alias — is version-controlled. This gives you:

- **History** — see exactly what you changed and when, and roll back mistakes.
- **Sync** — push to GitHub (or any Git remote) and pull on another machine.
- **Backup** — your dotfiles are safely stored in a remote repo, not just on one disk.

The daily workflow is simple: make changes, `gdf save` to commit, `gdf push` to upload. On another machine, `gdf pull` and `gdf apply`.

### Putting it all together

Here's the mental model in one picture:

```
┌────────────────────────────────────────────────────────────────────┐
│                         Your Home Directory                        │
│                                                                    │
│   ~/.gitconfig   ─► symlink ─► ~/.gdf/dotfiles/git/.gitconfig      │
│   ~/.vimrc       ─► symlink ─► ~/.gdf/dotfiles/vim/.vimrc          │
│   ~/.kube/config -► symlink ─► ~/.gdf/dotfiles/kubectl/kube/config │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────┐
│                          ~/.gdf/ (Git repo)                        │
│                                                                    │
│   apps/           ← app bundle YAML definitions                    │
│   dotfiles/       ← actual config file contents                    │
│   profiles/       ← profile definitions                            │
│   generated/      ← shell scripts (aliases, env vars)              │
│                                                                    │
│   All of this is committed to Git and synced across machines.      │
└────────────────────────────────────────────────────────────────────┘
```

Tools find their configs where they expect them (via symlinks). GDF keeps the real files in one place (inside `~/.gdf/`). Git makes it all portable.

Now that you understand the concepts, let's get GDF installed and start building your environment.


## Installing GDF

### Via Install Script (Recommended)

The easiest way to install GDF is using the official script:

```bash
curl -sfL https://raw.githubusercontent.com/rztaylor/GoDotFiles/main/scripts/install.sh | sh
```

### From Source

Requires Go 1.21+ installed on your machine.

```bash
go install github.com/rztaylor/GoDotFiles/cmd/gdf@latest
```

Verify the installation:

```bash
gdf --help
```

You should see the list of available commands, including core workflows like `init`, `add`, `track`, `apply`, `rollback`, `profile`, `alias`, `install`, `save`, `push`, `pull`, `sync`, and `status`.

---

## Initializing Your Dotfiles Repository

GDF stores everything under `~/.gdf/`. Let's create a fresh repository:

```bash
gdf init
```

You'll see output like:

```
✓ Created GDF directory at /home/you/.gdf
✓ Initialized git repository
✓ Created directory structure:
    apps/
    profiles/
    dotfiles/
```

### (Optional) Push to GitHub

If you want to back up your dotfiles and sync them across machines, create a remote repository now:

1. Go to [github.com/new](https://github.com/new) and create a **private** repository called `dotfiles` (no README, no `.gitignore`).
2. Link it to your local GDF repo and push:

```bash
git -C ~/.gdf remote add origin git@github.com:your-username/dotfiles.git
gdf save "Initial commit"
git -C ~/.gdf push -u origin main
```

You can skip this step and do it later — everything up to the [Saving and Syncing with Git](#saving-and-syncing-with-git) section works entirely offline.

> [!TIP]
> If you already have a dotfiles repo on GitHub, you can clone it instead of starting fresh:
> ```bash
> gdf init git@github.com:your-username/dotfiles.git
> ```
> This clones the repo into `~/.gdf/` so you can start applying profiles immediately.

### Shell Integration

GDF generates a shell script at `~/.gdf/generated/init.sh` containing your aliases, environment variables, functions, and startup init snippets. During `gdf init`, you'll be prompted to add the source line to your shell RC file automatically. If you prefer to do it manually, add this to your `~/.bashrc` or `~/.zshrc`:

```bash
[ -f ~/.gdf/generated/init.sh ] && source ~/.gdf/generated/init.sh
```

Reload your shell afterward:

```bash
source ~/.zshrc  # or ~/.bashrc
```

---

## Creating Your First Profile

Let's start with a **base** profile containing tools every machine should have:

```bash
gdf profile create base --description "Essential tools for every machine"
```

This creates a profile definition at `~/.gdf/profiles/base/profile.yaml`.

You can verify it exists:

```bash
gdf profile list
```

Output:

```
Profiles:
  base    Essential tools for every machine    (0 apps)
```

---

## Adding Apps and Tracking Dotfiles

Now let's populate the **base** profile with real tools.

### Git

Add Git as an app bundle and track your existing Git config:

```bash
# Create the app bundle
gdf add git --profile base

# Track your existing gitconfig
gdf track ~/.gitconfig --app git
```

The `track` command does two things:
1. **Copies** `~/.gitconfig` into `~/.gdf/dotfiles/git/` for version control.
2. **Replaces** the original with a **symlink** pointing back to the managed copy.

Your file is now safely stored in `~/.gdf/` and will follow you to any machine.

> [!NOTE]
> If you have a global `.gitignore` file, track that too:
> ```bash
> gdf track ~/.gitignore_global -a git
> ```

### Zsh

```bash
gdf add zsh -p base
gdf track ~/.zshrc -a zsh
```

### Vim

```bash
gdf add vim -p base
gdf track ~/.vimrc -a vim
```

### tmux

```bash
gdf add tmux -p base
gdf track ~/.tmux.conf -a tmux
```

### curl

Not every app needs dotfiles. Some are just packages:

```bash
gdf add curl -p base
```

### Verify your profile

```bash
gdf profile show base
```

Output:

```
Profile: base
Description: Essential tools for every machine

Apps (5):
  git      ✓ dotfiles
  zsh      ✓ dotfiles
  vim      ✓ dotfiles
  tmux     ✓ dotfiles
  curl
```

---

## Installing Apps Directly (New)

Sometimes you just want to install an app without manually creating a profile entry first. `gdf install` lets you do this, and it even learns how to install things if it doesn't know yet.

```bash
gdf install ripgrep
```

If GDF already knows how to install `ripgrep` (because a `ripgrep.yaml` app bundle exists with package instructions), it will just install it.

**Interactive Learning Mode:**

If you try to install an app that GDF *doesn't* have a package instruction for on your current OS (e.g., you're on Linux but `ripgrep.yaml` only has Homebrew instructions), GDF will ask you:

```text
$ gdf install ripgrep
App 'ripgrep' not found, creating new bundle...
⚠️  App 'ripgrep' is not in any profile.
Select a profile to add it to:
   1. base
   2. Create new profile...
   3. Skip (leave orphaned)
Select: 1
✓ Added 'ripgrep' to profile 'base'
❓ How do you install 'ripgrep' on linux?
   1. Package Manager (apt)
   2. Custom Script
   3. Skip (Manual/External)
Select [1-3]: 1
Enter package name for apt (default: ripgrep): ripgrep
✓ Updated app definition for 'ripgrep'
📦 Installing ripgrep via apt...
✅ Installed successfully
```

This updates the app bundle definitions permanently, so next time (or on another machine), it just works.

---

## Managing Shell Aliases

Aliases are one of the most powerful parts of GDF. Instead of scattering alias definitions across your `.bashrc` or `.zshrc`, GDF organises them by app and generates a single, clean shell script.

### Adding aliases

```bash
# Git aliases — auto-detected because a "git" app bundle exists
gdf alias add g git
gdf alias add gs "git status"
gdf alias add ga "git add"
gdf alias add gc "git commit"
gdf alias add gp "git push"
gdf alias add gl "git log --oneline -20"
gdf alias add gd "git diff"

# General productivity — stored as global (unassociated) aliases
# because there's no "ls" or "cd" app bundle
gdf alias add ll "ls -la"
gdf alias add la "ls -A"
gdf alias add ..  "cd .."
gdf alias add ... "cd ../.."
```

GDF's auto-detection checks whether the command's **first word** matches an **existing** app bundle. The `g` and `gs` aliases are associated with the `git` app bundle because a `git.yaml` app already exists. Aliases for commands like `ls` or `cd` (which have no app bundle) are stored as global aliases in `~/.gdf/aliases.yaml`.

You can always force an alias onto a specific app with the `-a` flag:

```bash
gdf alias add myalias "some complex pipeline" -a kubectl
```

### Listing aliases

```bash
gdf alias list
```

Output:

```
Aliases by app:

  git:
    g   = "git"
    gs  = "git status"
    ga  = "git add"
    gc  = "git commit"
    gp  = "git push"
    gl  = "git log --oneline -20"
    gd  = "git diff"

  (unassociated):
    ..  = "cd .."
    ... = "cd ../.."
    la  = "ls -A"
    ll  = "ls -la"
```

### Removing an alias

```bash
gdf alias remove ..
```

This searches all app bundles and global aliases to find and remove the alias.

---

## Building a DevOps/SRE Profile

Now let's create a profile for SRE/DevOps work:

```bash
gdf profile create sre --description "SRE and DevOps tooling"
```

### Adding SRE tools

```bash
# Container orchestration
gdf add kubectl -p sre
gdf add docker -p sre

# Infrastructure as Code
gdf add terraform -p sre

# Cloud CLIs
gdf add aws-cli -p sre

# Monitoring
gdf add jq -p sre
```

### Tracking SRE dotfiles

Track your Kubernetes config and AWS credentials:

```bash
gdf track ~/.kube/config -a kubectl
gdf track ~/.aws/config -a aws-cli --secret
```

> [!CAUTION]
> **Sensitive files:** Be careful tracking files containing secrets. Always use the `--secret` flag (e.g., `gdf track ~/.aws/credentials --secret`) to ensure they are added to `.gitignore` automatically. Never commit secrets to your repository.

> [!TIP]
> **Why track gitignored files?**
> Even though the content won't be synced to Git, tracking ensures:
> 1. **Central location**: All your config files live in `~/.gdf` for easier backup.
> 2. **Symlink management**: GDF maintains the link for you.
> 3. **Inventory**: Your `app.yaml` records that this file is part of your configuration.

### Adding SRE aliases

```bash
gdf alias add k kubectl -a kubectl
gdf alias add kgp "kubectl get pods" -a kubectl
gdf alias add kgs "kubectl get svc" -a kubectl
gdf alias add kgn "kubectl get nodes" -a kubectl
gdf alias add kns "kubectl config set-context --current --namespace" -a kubectl
gdf alias add tf terraform -a terraform
gdf alias add tfi "terraform init" -a terraform
gdf alias add tfp "terraform plan" -a terraform
gdf alias add tfa "terraform apply" -a terraform
gdf alias add dps "docker ps" -a docker
gdf alias add dimg "docker images" -a docker
```

### Verify the SRE profile

```bash
gdf profile show sre
```

```
Profile: sre
Description: SRE and DevOps tooling

Apps (5):
  kubectl     ✓ dotfiles
  docker
  terraform
  aws-cli     ✓ dotfiles
  jq
```

---

## Building a Programming Profile

Let's create a profile for your development languages:

```bash
gdf profile create programming --description "Programming languages and dev tools"
```

### Go

```bash
gdf add go -p programming
```

If you have custom Go environment settings, you can track them. Most Go developers set environment variables via their shell, so let's edit the app bundle YAML directly to add environment variables (we'll cover YAML editing in [Editing App Bundles by Hand](#editing-app-bundles-by-hand-yaml)):

```bash
# For now, just add the app and useful aliases
gdf alias add gotest "go test ./..." -a go
gdf alias add govet "go vet ./..." -a go
gdf alias add gobuild "go build ./..." -a go
```

### Python

```bash
gdf add python -p programming
```

Track your Python configuration files:

```bash
# pip configuration
gdf track ~/.config/pip/pip.conf -a python

# If you use pylint or flake8
gdf track ~/.config/flake8 -a python
```

Add Python aliases:

```bash
gdf alias add py python3 -a python
gdf alias add pip "python3 -m pip" -a python
gdf alias add venv "python3 -m venv" -a python
gdf alias add activate "source .venv/bin/activate" -a python
```

### Additional dev tools

```bash
gdf add make -p programming
gdf add ripgrep -p programming
gdf add fzf -p programming
```

Add handy search aliases:

```bash
gdf alias add rg ripgrep -a ripgrep
```

### Verify

```bash
gdf profile show programming
```

```
Profile: programming
Description: Programming languages and dev tools

Apps (6):
  go
  python    ✓ dotfiles
  make
  ripgrep
  fzf
```

---

## Meta-Profiles: Grouping Apps

As your setup grows, you might find yourself adding the same groups of apps to multiple profiles. For example, `node` and `npm` almost always go together. `kubectl`, `helm`, and `docker` are often a set.

GDF allows you to create **Meta-Apps**—special recipes that exist solely to group other apps together.

### Example: Backend Developer Suite

Imagine you want a single item that installs Go, Docker, Kubectl, and Terraform. You can create a "Meta-App" recipe called `backend-dev`:

1.  Create `~/.gdf/apps/backend-dev.yaml`:

    ```yaml
    name: backend-dev
    description: "Backend Development Suite"
    kind: App/v1  # Uses App schema but acts as a group

    dependencies:
      - go
      - docker
      - kubectl
      - terraform
    ```

2.  Add it to your profile:

    ```bash
    gdf add backend-dev -p programming
    ```

Now, when you run `gdf apply programming`, GDF sees `backend-dev`, checks its dependencies, and automatically ensures `go`, `docker`, `kubectl`, and `terraform` are also applied—even if you didn't list them explicitly in the profile.

This keeps your profiles clean and composable. You can have a `frontend-dev` meta-app (Node, NPM, GitHub CLI) and a `backend-dev` meta-app, and just mix and match them.

---

## Applying Profiles

Now that your profiles are defined, let's apply them to your current machine.

### Dry run first

Always preview what GDF will do before making changes:

```bash
gdf apply --dry-run base sre programming
```

This shows you which packages would be installed, which dotfiles would be symlinked, and which aliases would be generated — without making any changes.

### Apply for real

```bash
gdf apply base sre programming
```

GDF performs these operations in order:

1. **Resolves profile dependencies** — if a profile includes other profiles, they're processed in the right order.
2. **Resolves app dependencies** — apps are ordered topologically.
3. **Installs packages** — uses the appropriate package manager for your OS (Homebrew on macOS, apt on Ubuntu/Debian, dnf on Fedora).
4. **Links dotfiles** — creates symlinks from your home directory to `~/.gdf/dotfiles/`.
5. **Generates shell integration** — writes `~/.gdf/generated/init.sh` with all your aliases, environment variables, functions, and startup init snippets.
6. **Scans for high-risk commands** — warns and asks for confirmation before proceeding when risky hook/script patterns are detected.
7. **Logs operations** — saves a timestamped log to `~/.gdf/.operations/` for rollback.
8. **Captures history snapshots** — stores pre-change copies under `~/.gdf/.history/` before destructive replacement operations.
9. **Updates state** — records which profiles were applied to `~/.gdf/state.yaml` (local only).

### Check status

```bash
gdf status
```

Output:

```
Applied Profiles:
  ✓ base (5 apps) - applied just now
  ✓ sre (5 apps) - applied just now
  ✓ programming (6 apps) - applied just now

Apps (16 total):
  git, zsh, vim, tmux, curl, kubectl, docker, terraform, aws-cli,
  jq, go, python, make, ripgrep, fzf

Last applied: 2026-02-11 21:00:00
```

Reload your shell to pick up the generated aliases:

```bash
source ~/.zshrc  # or: gdf shell reload
```

Try out your new aliases:

```bash
gs        # → git status
k         # → kubectl (if installed)
tf        # → terraform
py        # → python3
```

### Rollback and historical snapshots

If an apply introduces a bad config change, you can rollback immediately:

```bash
gdf rollback
```

If you need to restore a specific file and choose from multiple dated snapshots:

```bash
gdf rollback --target ~/.zshrc --choose-snapshot
```

Snapshot storage uses `~/.gdf/.history/` and is quota-based. You can tune it in `~/.gdf/config.yaml`:

```yaml
history:
  max_size_mb: 512
```

---

## Saving and Syncing with Git

Everything is local so far. Let's back it all up to a Git remote.

### Create a remote repository

Go to GitHub (or GitLab, Bitbucket, etc.) and create a new **private** repository called `dotfiles`.

### Add the remote

```bash
git -C ~/.gdf remote add origin git@github.com:your-username/dotfiles.git
```

### Save and push

```bash
# Stage and commit all changes
gdf save "Initial setup: base, sre, and programming profiles"

# Push to remote (first time requires setting upstream)
git -C ~/.gdf push -u origin main
```

For subsequent pushes, you can simply use:

```bash
gdf push
```

### The sync shortcut

`gdf sync` combines pull + commit + push into one command. It's the recommended way to stay up to date:

```bash
gdf sync
```

This is equivalent to:

```bash
gdf pull              # Get changes from remote
gdf save "Update"     # Commit any local changes
gdf push              # Push to remote
```

---

## Setting Up a Second Machine

This is where GDF shines. You've got a new work server (or a fresh laptop) and you want your full environment running in minutes.

### Step 1: Install GDF

```bash
curl -sfL https://raw.githubusercontent.com/rztaylor/GoDotFiles/main/scripts/install.sh | sh
```

### Step 2: Clone your dotfiles

```bash
gdf init git@github.com:your-username/dotfiles.git
```

This clones your entire `~/.gdf/` repository, including all app bundles, profiles, and tracked dotfiles.

### Step 3: Apply the profiles you need

On a work server, you might only want base tools and SRE tooling — skip the programming profile:

```bash
# Preview first
gdf apply --dry-run base sre

# Apply
gdf apply base sre
```

On your personal laptop, apply everything:

```bash
gdf apply base sre programming
```

### Step 4: Reload your shell

```bash
source ~/.zshrc  # or ~/.bashrc
```

That's it. Your aliases, dotfiles, and (where possible) packages are all in place. Your `~/.gitconfig`, `~/.vimrc`, `~/.tmux.conf`, and `~/.kube/config` are all symlinked from the managed copies in `~/.gdf/dotfiles/`.

---

## Day-to-Day Workflow

Once your setup is running, here's the typical workflow:

### Making changes

```bash
# Track a new dotfile
gdf track ~/.ssh/config -a ssh

# Add a new alias
gdf alias add dc docker-compose

# Add a new tool
gdf add helm -p sre

# Save and sync
gdf sync
```

### On another machine

```bash
# Pull the latest changes
gdf sync

# Re-apply profiles to pick up new apps/dotfiles
gdf apply base sre
```

### Checking what's applied

```bash
gdf status
```

### Full worked example: adding a new tool

Let's say you've started using **Helm** for Kubernetes package management. Here's the complete workflow:

```bash
# 1. Add Helm to the SRE profile
gdf add helm -p sre

# 2. If you have a Helm config, track it
gdf track ~/.config/helm/repositories.yaml -a helm

# 3. Add useful aliases
gdf alias add h helm -a helm
gdf alias add hls "helm list" -a helm
gdf alias add hup "helm upgrade" -a helm

# 4. Apply to get everything linked and generated
gdf apply sre

# 5. Save and sync to all machines
gdf save "Added Helm with config and aliases"
gdf push

# 6. On other machines:
gdf pull
gdf apply sre
```

---

## Editing App Bundles by Hand (YAML)

The CLI covers common operations, but you can always edit the YAML files directly for full control. App bundles live in `~/.gdf/apps/`.

### Example: kubectl app bundle

Open `~/.gdf/apps/kubectl.yaml` in your editor:

```yaml
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
    kgs: kubectl get svc
    kgn: kubectl get nodes
    kns: kubectl config set-context --current --namespace
  env:
    KUBECONFIG: ~/.kube/config
  completions:
    bash: kubectl completion bash
    zsh: kubectl completion zsh
```

### Example: Go app bundle with environment variables

`~/.gdf/apps/go.yaml`:

```yaml
name: go
description: Go programming language

package:
  brew: go
  apt: golang
  dnf: golang

shell:
  aliases:
    gotest: go test ./...
    govet: go vet ./...
    gobuild: go build ./...
  env:
    GOPATH: ~/go
    GOBIN: ~/go/bin
```

### Example: Python app bundle with pip config

`~/.gdf/apps/python.yaml`:

```yaml
name: python
description: Python 3 programming language

package:
  brew: python@3
  apt: python3
  dnf: python3

dotfiles:
  - source: pip/pip.conf
    target: ~/.config/pip/pip.conf
  - source: flake8
    target: ~/.config/flake8

shell:
  aliases:
    py: python3
    pip: python3 -m pip
    venv: python3 -m venv
    activate: source .venv/bin/activate
  env:
    PYTHONDONTWRITEBYTECODE: "1"
```

### Example: Terraform with dependencies

`~/.gdf/apps/terraform.yaml`:

```yaml
name: terraform
description: Infrastructure as Code

dependencies:
  - aws-cli    # Ensure AWS CLI is installed first

package:
  brew: terraform
  apt:
    name: terraform
    repo: https://apt.releases.hashicorp.com
    key: https://apt.releases.hashicorp.com/gpg
  dnf:
    name: terraform

shell:
  aliases:
    tf: terraform
    tfi: terraform init
    tfp: terraform plan
    tfa: terraform apply
  completions:
    bash: terraform -install-autocomplete
```

### Profile includes

You can make a **work** profile that automatically includes `base` and `sre`:

Edit `~/.gdf/profiles/work/profile.yaml`:

```yaml
name: work
description: Work environment (includes base + sre)
includes:
  - base
  - sre
```

Now `gdf apply work` automatically applies `base` and `sre` first.

---

## Tips and Best Practices

### Organise profiles by context, not by tool type

Instead of profiles like `cli-tools` and `gui-tools`, think in terms of **when and where** you need them:

| Profile | Purpose | Machines |
|---------|---------|----------|
| `base` | Essential tools for every system | All |
| `programming` | Languages and dev tools | Laptops, dev servers |
| `sre` | Kubernetes, Terraform, cloud CLIs | Work machines |
| `work` | Includes base + sre, adds work-specific tools | Work machines |
| `personal` | Personal tools and configs | Personal machines |

### Use `--dry-run` before every apply

```bash
gdf apply --dry-run base sre
```

This habit will prevent surprises, especially when you `pull` changes made on another machine.

### Sync frequently

```bash
# Make it a habit: end of each session
gdf sync
```

This keeps all your machines up to date and prevents merge conflicts.

### Keep secrets out of Git

Never commit credentials. Mark sensitive files as secrets in YAML:

```yaml
dotfiles:
  - source: aws/credentials
    target: ~/.aws/credentials
    secret: true              # Auto-added to .gitignore
```

For files like `~/.aws/credentials` or `~/.ssh/id_rsa`, either:
- Use the `secret: true` flag (tracks the path, but gitignores the content).
- Don't track them at all — manage them separately with a secrets manager.

### Use descriptive save messages

```bash
# Good
gdf save "Added helm config and SRE aliases"
gdf save "Tracked new python flake8 config"

# Less useful
gdf save
gdf save "update"
```

### Check status regularly

```bash
gdf status
```

This tells you which profiles are applied, how many apps are active, and when they were last applied.

---

## What's Next?

You've covered the core GDF workflow. Next useful areas to explore:

- **Library Recipes** — Browse with `gdf library list` and inspect with `gdf library describe <recipe>`.
- **Conditional dotfiles** — Use `dotfiles[].when` and platform-specific targets in app YAML for per-host/per-OS behavior.
- **Recovery workflows** — Use `gdf rollback` and targeted snapshot restore when testing risky changes.
- **Upcoming commands** — Future phases include richer diagnostics and workflows like `gdf doctor`, `gdf validate`, and interactive setup tools.

For complete command details, see the [CLI Reference](../reference/cli.md). For YAML syntax, see the [YAML Schema Reference](../reference/yaml-schemas.md).
