---
description: How to cut a new release of GDF
---

1.  **Prepare Changelog**:
    -   Open `CHANGELOG.md`
    -   Move `[Unreleased]` items to a new version header (e.g., `## [0.6.0] - 2024-03-20`)
    -   Update the comparison links at the bottom if applicable

2.  **Commit Changes**:
    -   Extract release notes for the commit message:
        ```bash
        go run scripts/extract_release_notes.go > release_notes.txt
        ```
    -   Commit with the release notes:
        ```bash
        git commit -a -F release_notes.txt
        rm release_notes.txt
        ```

3.  **Run Release**:
    -   Run the release command with the new version tag:
    ```bash
    make release VERSION=v0.6.0
    ```
    -   This will verify the git state, tag the commit, and push to origin.

4.  **Verify**:
    -   Check the [GitHub Actions](https://github.com/rztaylor/GoDotFiles/actions) tab to ensure the release workflow runs successfully.
