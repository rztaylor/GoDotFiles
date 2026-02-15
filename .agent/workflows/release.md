---
description: How to cut a new release of GDF
---

1.  **Prepare Changelog**:
    -   Open `CHANGELOG.md`
    -   Move `[Unreleased]` items to a new version header (e.g., `## [0.6.0] - 2024-03-20`)
    -   Keep entries user-facing (behavior and outcomes), not internal implementation details.
    -   Update comparison links at the bottom if applicable.

2.  **Preflight Checks**:
    -   Ensure working tree is clean before release tagging:
        ```bash
        git status --short
        ```
    -   Ensure the release tag does not already exist:
        ```bash
        git tag --list vX.Y.Z
        ```
    -   Optional: also check remote tag:
        ```bash
        git ls-remote --tags origin vX.Y.Z
        ```

3.  **Commit Changes**:
    -   Extract release notes for the commit message (write temp file outside repo):
        ```bash
        go run ./scripts/extract-release-notes > /tmp/release_notes.txt
        ```
    -   Commit with release notes:
        ```bash
        git add -A
        git commit -F /tmp/release_notes.txt
        ```

4.  **Run Release**:
    -   Run the release command with the new version tag:
    ```bash
    make release VERSION=vX.Y.Z
    ```
    -   This will verify the git state, tag the commit, and push to origin.

5.  **Failure Recovery (Tag Created, Push Failed)**:
    -   If release command fails after creating local tag (e.g., network failure), do not recreate the tag.
    -   Push existing tag directly:
    ```bash
    git push origin vX.Y.Z
    ```

6.  **Verify**:
    -   Check the [GitHub Actions](https://github.com/rztaylor/GoDotFiles/actions) tab to ensure the release workflow runs successfully.
