## Why

The current pacman implementation spawns multiple subprocesses per package (pacman -Qip, pacman -Sip) to check if packages exist in local/sync DBs or AUR. With many packages, this creates significant overhead. Using the Jguer/dyalpm library provides direct libalpm access for batch queries, eliminating subprocess overhead while maintaining the batched AUR HTTP calls.

## What Changes

- **Add dyalpm dependency**: Integrate Jguer/dyalpm library for direct libalpm access
- **Batch local DB check**: Use `localDB.PkgCache()` to check all packages at once instead of per-package `pacman -Qip`
- **Batch sync DB check**: Use `syncDBs[i].PkgCache()` to check all sync repos at once instead of per-package `pacman -Sip`
- **Enhance PackageInfo**: Add `Installed bool` field to track if package is already installed
- **New algorithm**: Implement unified package resolution flow:
  1. Batch check local DB for all packages
  2. Batch check sync DBs for remaining packages
  3. Batch query AUR for non-found packages
  4. Track installed status throughout
  5. Perform sync operations with proper marking
  6. Output summary of changes

## Capabilities

### New Capabilities

- `batch-package-resolution`: Unified algorithm that batch-resolves packages from local DB → sync DBs → AUR with proper installed tracking
- `dry-run-simulation`: Shows exact packages that would be installed/removed without making changes

### Modified Capabilities

- None - this is a pure optimization with no behavior changes visible to users

## Impact

- **Code**: `pkg/pacman/pacman.go` - refactored to use dyalpm
- **Dependencies**: Add Jguer/dyalpm to go.mod
- **APIs**: `ValidatePackage()` signature changes (returns installed status)
- **Performance**: O(n) subprocess calls → O(1) for local/sync DB checks