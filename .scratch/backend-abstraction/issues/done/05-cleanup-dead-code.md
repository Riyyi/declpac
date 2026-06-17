# 05 - Clean Up Dead Code and Update Result Types

Status: ready-for-agent

## Description

Remove `sync.Result` from `pkg/pacman/sync/sync.go` — it overlaps with `output.Result` and is never used as a return type.

Verify that `output.Result` is the single result type used across the codebase. No changes to its fields (stays flat strings).

## Acceptance Criteria

- `sync.Result` struct removed
- No references to `sync.Result` remain
- `output.Result` is the only result struct