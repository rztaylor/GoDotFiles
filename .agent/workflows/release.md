---
description: How to cut a new release of GDF
---

1.  **Prepare Changelog**:
    -   Open `CHANGELOG.md`
    -   Move `[Unreleased]` items to a new version header (e.g., `## [0.6.0] - 2024-03-20`)
    -   Update the comparison links at the bottom if applicable

2.  **Commit Changes**:
    -   `git commit -am "chore: prepare release v0.6.0"`

3.  **Run Release**:
    -   Run the release command with the new version tag:
    ```bash
    make release VERSION=v0.6.0
    ```
    -   This will verify the git state, tag the commit, and push to origin.

4.  **Verify**:
    -   Check the [GitHub Actions](https://github.com/rztaylor/GoDotFiles/actions) tab to ensure the release workflow runs successfully.
