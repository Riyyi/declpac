# 02 - Refactor Pacman to Implement Backend

Status: ready-for-agent

## Description

Make `pkg/pacman/pacman.go` implement the `Backend` interface. Move `pkg/fetch/` to `pkg/pacman/fetch/`. Wire `Resolve`, `List`, `ListExplicit`, `ListOrphans`, `RefreshDB`, `InstallPackages`, `RemoveOrphans`, `MarkAs`, and `Prune` as methods on a `Pacman` struct.

Key changes:
- Create `Pacman` struct that holds `*fetch.Fetcher` (created in constructor, closed on demand)
- `Resolve` wraps `fetcher.Resolve`, converts `fetch.PackageInfo` → `backend.PackageInfo`
- `InstallPackages` receives `[]backend.PackageInfo`, filters by `Source=="repo"` for `sync.SyncPackages`, then loops `Source=="aur"` for `sync.InstallAUR`
- `Prune` contains the current 3-step pacman flow (markAllAsDeps, markStateExplicit, cleanupOrphans)
- `RefreshDB` checks `read.DBFreshness()` internally, calls `sync.RefreshDB` only if stale
- `MarkAs` delegates to `sync.MarkAs`
- Update all imports from `pkg/fetch` to `pkg/pacman/fetch`

## Acceptance Criteria

- `*pacman.Pacman` satisfies `backend.Backend` interface
- All existing pacman functionality preserved
- `pkg/fetch/` deleted, code lives under `pkg/pacman/fetch/`
- No changes to external behavior