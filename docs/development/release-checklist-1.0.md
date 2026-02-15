# 1.0 Release Checklist

Use this checklist before tagging the public `1.0.0` release.

## Safety and Recovery

- [ ] `gdf apply --dry-run` works for representative profiles.
- [ ] `gdf recover rollback` restores from latest operation logs.
- [ ] Destructive lifecycle flows provide explicit previews before action (`app remove --uninstall`, `profile delete --dry-run`).
- [ ] Profile delete mode flags are validated as mutually exclusive (`--migrate-to-default`, `--purge`, `--leave-dangling`).

## Determinism and Automation

- [ ] Interactive paths honor `--non-interactive` and `--yes` contracts.
- [ ] Exit code contracts match documented values in `docs/reference/compatibility.md`.
- [ ] JSON output commands continue to emit machine-readable output without prompt text contamination.

## Docs and Contract Lock-in

- [ ] `docs/reference/cli.md` reflects shipped commands and flags.
- [ ] `docs/reference/yaml-schemas.md` reflects `v1` stability contract.
- [ ] `docs/reference/compatibility.md` is current and matches behavior.
- [ ] `CHANGELOG.md` `[Unreleased]` contains user-facing notes for all shipped changes.

## Validation

- [ ] `go test ./...` passes.
- [ ] `golangci-lint run` passes.
- [ ] Release smoke test performed on at least one macOS and one Linux environment.
