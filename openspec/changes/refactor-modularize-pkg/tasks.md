# Tasks: Refactor pkg Into Modular Packages

## Phase 1: Create pkg/fetch

- [ ] 1.1 Create `pkg/fetch/fetch.go`
- [ ] 1.2 Move `AURResponse`, `AURPackage`, `PackageInfo` structs to fetch
- [ ] 1.3 Move `buildLocalPkgMap()` to fetch as `Fetcher.buildLocalPkgMap()`
- [ ] 1.4 Move `checkSyncDBs()` to fetch as `Fetcher.checkSyncDBs()`
- [ ] 1.5 Move `resolvePackages()` to fetch as `Fetcher.Resolve()`
- [ ] 1.6 Move AUR cache methods (`ensureAURCache`, `fetchAURInfo`) to fetch
- [ ] 1.7 Add `New()` and `Close()` to Fetcher
- [ ] 1.8 Add `ListOrphans()` to Fetcher

## Phase 2: Refactor pkg/pacman

- [ ] 2.1 Remove from pacman.go (now in fetch):
  - `buildLocalPkgMap()`
  - `checkSyncDBs()`
  - `resolvePackages()`
  - `ensureAURCache()`
  - `fetchAURInfo()`
  - `AURResponse`, `AURPackage`, `PackageInfo` structs
- [ ] 2.2 Remove `IsDBFresh()` and `SyncDB()` (use validation instead)
- [ ] 2.3 Update imports in pacman.go to include fetch package
- [ ] 2.4 Update `Sync()` to use `fetch.Fetcher` for resolution
- [ ] 2.5 Update `DryRun()` to call `fetcher.ListOrphans()` instead of duplicate call
- [ ] 2.6 Update `CleanupOrphans()` to call `fetcher.ListOrphans()` instead of duplicate call

## Phase 3: Clean Up Validation

- [ ] 3.1 Keep `validation.CheckDBFreshness()` as-is
- [ ] 3.2 Remove any remaining DB freshness duplication

## Phase 4: Verify

- [ ] 4.1 Run tests (if any exist)
- [ ] 4.2 Build: `go build ./...`
- [ ] 4.3 Verify CLI still works: test dry-run, sync, orphan cleanup