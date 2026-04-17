# Proposal: Refactor pkg Into Modular Packages

## Summary

Split monolithic `pkg/pacman/pacman.go` into focused packages: `fetch` (package resolution), `validation` (DB checks), and keep `pacman` for write actions only. Also deduplicate orphan detection logic.

## Motivation

`pacman.go` is 645 lines doing too much:
- Package resolution (local + sync DBs + AUR)
- DB freshness checks
- Pacman write operations (sync, mark, clean)
- Orphan listing/cleanup

This violates single responsibility. Hard to test, reason about, or reuse. Also has duplication:
- `validation.CheckDBFreshness()` and `pacman.IsDBFresh()` both check DB freshness
- `listOrphans()` and `CleanupOrphans()` duplicate orphan detection

## Scope

- Extract package resolution to new `pkg/fetch/`
- Move DB freshness to `pkg/validation/` (keep `CheckDBFreshness()`)
- Keep only write actions in `pkg/pacman/`
- Deduplicate orphan logic: one function for listing, reuse in cleanup and dry-run

## Out of Scope

- No new features
- No API changes to CLI
- No changes to `pkg/output/`, `pkg/merge/`, `pkg/input/`