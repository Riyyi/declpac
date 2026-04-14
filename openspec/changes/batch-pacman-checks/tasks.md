## 1. Setup

- [x] 1.1 Add `github.com/Jguer/dyalpm` to go.mod
- [x] 1.2 Run `go mod tidy` to fetch dependencies

## 2. Core Refactoring

- [x] 2.1 Update `PackageInfo` struct to add `Installed bool` field
- [x] 2.2 Create `Pac` struct with `alpm.Handle` instead of just aurCache
- [x] 2.3 Implement `NewPac()` that initializes alpm handle and local/sync DBs

## 3. Package Resolution Algorithm

- [x] 3.1 Implement `buildLocalPkgMap()` - iterate localDB.PkgCache() to create lookup map
- [x] 3.2 Implement `checkSyncDBs()` - iterate each sync DB's PkgCache() to find packages
- [x] 3.3 Implement `resolvePackages()` - unified algorithm:
  - Step 1: Check local DB for all packages (batch)
  - Step 2: Check sync DBs for remaining packages (batch per repo)
  - Step 3: Batch query AUR for remaining packages
  - Step 4: Return error if any package unfound
  - Step 5: Track installed status from local DB

## 4. Sync and DryRun Integration

- [x] 4.1 Refactor `Sync()` function to use new resolution algorithm
- [x] 4.2 Refactor `DryRun()` function to use new resolution algorithm
- [x] 4.3 Preserve AUR batched HTTP calls (existing `fetchAURInfo`)
- [x] 4.4 Preserve orphan cleanup logic (`CleanupOrphans()`)

## 5. Marking Operations

- [x] 5.1 Keep `MarkExplicit()` for marking state packages
- [x] 5.2 After sync, run `pacman -D --asdeps` on ALL installed packages (simplifies tracking)
- [x] 5.3 After deps marking, run `pacman -D --asexplicit` on collected state packages (overrides deps)
- [x] 5.4 Skip marking operations in dry-run mode

## 6. Cleanup and Output

- [x] 6.1 Remove subprocess-based `ValidatePackage()` implementation
- [x] 6.2 Remove subprocess-based `GetInstalledPackages()` implementation
- [x] 6.3 Update output summary to show installed/removed counts
- [x] 6.4 In dry-run mode, populate `ToInstall` and `ToRemove` lists

## 7. Testing

- [ ] 7.1 Test with packages in local DB only
- [ ] 7.2 Test with packages in sync DBs only
- [ ] 7.3 Test with AUR packages
- [ ] 7.4 Test with missing packages (should error)
- [ ] 7.5 Test dry-run mode output
- [ ] 7.6 Test orphan detection and cleanup