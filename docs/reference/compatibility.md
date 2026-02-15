# CLI Compatibility Policy (1.x)

This document defines compatibility guarantees for the public `1.x` series.

## Stability Guarantees

For all `1.x` releases:
- Existing commands keep their behavior unless explicitly documented as deprecated.
- Existing flags keep their meaning and defaults.
- Interactive command paths continue to honor `--non-interactive` and `--yes` according to command safety.
- Stable exit code contracts remain unchanged for automation:
  - `0`: success
  - `1`: runtime error
  - `2`: health issues
  - `3`: fix failure
  - `4`: non-interactive stop / confirmation required

## Deprecation Policy

When a command or flag must change in `1.x`:
1. Keep the old spelling as an alias for at least one minor release.
2. Emit a clear deprecation message with replacement guidance.
3. Update `docs/reference/cli.md` and `CHANGELOG.md` in the same release.
4. Remove the deprecated path only in a later minor release after the deprecation window.

## Non-Goals for 1.x

The following may be added in `1.x` without breaking compatibility:
- New commands and flags
- New JSON fields (additive only)
- New optional YAML fields under existing `v1` kinds

The following are considered breaking and are deferred to a future major version:
- Renaming/removing commands or flags without deprecation window
- Changing existing flag defaults in ways that alter behavior
- Reinterpreting existing YAML `v1` fields incompatibly
