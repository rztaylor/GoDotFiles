# Release Process

This document outlines the steps to release a new version of GDF.

## Prerequisites

-   Write access to the `rztaylor/GoDotFiles` repository.
-   `make` installed locally.
-   `git` configured.

## Release Steps

1.  **Update Changelog**:
    -   Edit `CHANGELOG.md`.
    -   Move all items from `[Unreleased]` to a new version section (e.g., `## [0.6.0] - YYYY-MM-DD`).
    -   Commit these changes: `git commit -am "chore: prepare release v0.6.0"`

2.  **Tag and Push**:
    -   Run the release target with the version number:
        ```bash
        make release VERSION=v0.6.0
        ```
    -   This command will:
        -   Verify the version format `vX.Y.Z`.
        -   Ensure git status is clean.
        -   Create a git tag `v0.6.0`.
        -   Push the tag to `origin`.

3.  **Verify GitHub Action**:
    -   Go to [GitHub Actions](https://github.com/rztaylor/GoDotFiles/actions).
    -   Watch the "Release" workflow.
    -   Once completed, a new release will appear on the [Releases Page](https://github.com/rztaylor/GoDotFiles/releases).

## Automated Artifacts

The GitHub Action uses [GoReleaser](https://goreleaser.com/) to automatically:
-   Build binaries for Linux (amd64, arm64) and macOS (amyl4, arm64).
-   Create `.tar.gz` and `.zip` archives.
-   Generate `checksums.txt`.
-   Draft the GitHub Release with the changelog.
