# Tasks: Refactor pkg Into Modular Packages

## Phase 1: Create pkg/fetch

- [x] 1.1 Create `pkg/fetch/fetch.go`
- [x] 1.2 Move `AURResponse`, `AURPackage`, `PackageInfo` structs to fetch
- [x] 1.3 Move `buildLocalPkgMap()` to fetch as `Fetcher.buildLocalPkgMap()`
- [x] 1.4 Move `checkSyncDBs()` to fetch as `Fetcher.checkSyncDBs()`
- [x] 1.5 Move `resolvePackages()` to fetch as `Fetcher.Resolve()`
- [x] 1.6 Move AUR cache methods (`ensureAURCache`, `fetchAURInfo`) to fetch
- [x] 1.7 Add `New()` and `Close()` to Fetcher
- [x] 1.8 Add `ListOrphans()` to Fetcher

## Phase 2: Refactor pkg/pacman

- [x] 2.1 Remove from pacman.go (now in fetch):
  - `buildLocalPkgMap()`
  - `checkSyncDBs()`
  - `resolvePackages()`
  - `ensureAURCache()`
  - `fetchAURInfo()`
  - `AURResponse`, `AURPackage`, `PackageInfo` structs
- [x] 2.2 Remove `IsDBFresh()` and `SyncDB()` (use validation instead)
- [x] 2.3 Update imports in pacman.go to include fetch package
- [x] 2.4 Update `Sync()` to use `fetch.Fetcher` for resolution
- [x] 2.5 Update `DryRun()` to call `fetcher.ListOrphans()` instead of duplicate call
- [x] 2.6 Update `CleanupOrphans()` to call `fetcher.ListOrphans()` instead of duplicate call

## Phase 3: Clean Up Validation

- [x] 3.1 Keep `validation.CheckDBFreshness()` as-is
- [x] 3.2 Remove any remaining DB freshness duplication

## Phase 4: Verify

- [x] 4.1 Run tests (if any exist)
- [x] 4.2 Build: `go build ./...`
- [x] 4.3 Verify CLI still works: test dry-run, sync, orphan cleanup