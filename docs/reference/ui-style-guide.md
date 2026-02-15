# UI Style Guide

This guide defines how GDF should present human-readable CLI output.

## Goals

- Fast scanability for interactive use.
- Professional tone suitable for daily engineering workflows.
- Consistent structure across commands.
- Deterministic non-interactive behavior for automation.

## Output Structure

Preferred section order:

1. Result status line
2. `Summary`
3. `Details` (when needed)
4. `Next Step`

Use section headings and whitespace deliberately. Avoid dense, unstructured blocks.

## Status Markers

Use only restrained status markers:

- `✓` success
- `✗` error
- `!` warning

Avoid decorative emoji.

## Color Policy

- Support `--color auto|always|never`.
- Respect `NO_COLOR`.
- In `auto` mode, colorize only for interactive terminals.
- Use color to reinforce structure and status, not as decoration.

Recommended usage:

- Section headings: blue (or equivalent emphasis).
- Key labels in `key: value` lines: cyan (or equivalent emphasis).
- Success/warn/error markers: green/yellow/red.

## Key/Value Formatting

For summaries and diagnostics, prefer aligned `key: value` lines:

```text
Summary
  Profile:   work
  Apps:      9
  Dotfiles:  34
```

Keys should align vertically for quick scanning.

## Verbose Output

- Default output should remain concise and actionable.
- `-v` should add deeper context (extra sections, itemized details).
- Do not require extra verbosity for basic understanding.

## Guided Defaults

When required context is missing in interactive terminals (for example, no profile specified and multiple profiles exist), default to guided prompts instead of hard failures.

In `--non-interactive` mode:

- Never prompt.
- Fail with explicit, actionable remediation steps.

