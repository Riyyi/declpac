## Context

Currently, `pkg/pacman/pacman.go` uses subprocess calls to query pacman for package existence:
- `pacman -Qip <pkg>` to check local DB (per package)
- `pacman -Sip <pkg>` to check sync repos (per package)

For n packages, this spawns 2n subprocesses (up to ~300 for typical package lists). Each subprocess has fork/exec overhead, making this the primary performance bottleneck.

The AUR queries are already batched (single HTTP POST with all package names), which is the desired pattern.

## Goals / Non-Goals

**Goals:**
- Eliminate subprocess overhead for local/sync DB package lookups
- Maintain batched AUR HTTP calls (single request per batch)
- Track installed status per package in PackageInfo
- Provide dry-run output showing exact packages to install/remove
- Handle orphan cleanup correctly

**Non-Goals:**
- Parallel AUR builds (still sequential)
- Custom pacman transaction handling (use system pacman)
- Repository configuration changes
- Package download/compile optimization

## Decisions

### 1. Use Jguer/dyalpm for DB access

**Decision**: Use `github.com/Jguer/dyalpm` library instead of spawning subprocesses.

**Rationale**:
- Direct libalpm access (same backend as pacman)
- Already Go-native with proper type safety
- Supports batch operations via `GetPkgCache()` and `PkgCache()` iterators

**Alternatives considered**:
- Parse `pacman -Qs` output - fragile, still subprocess-based
- Write custom libalpm bindings - unnecessary effort

### 2. Single-pass package resolution algorithm

**Decision**: Process all packages through local DB → sync DBs → AUR in one pass.

```
For each package in collected state:
  1. Check local DB (batch lookup) → if found, mark Installed=true
  2. If not local, check all sync DBs (batch lookup per repo)
  3. If not in sync, append to AUR batch

Batch query AUR with all remaining packages
Throw error if any package not found in local/sync/AUR

Collect installed status from local DB
(Perform sync operations - skip in dry-run)
(Mark ALL currently installed packages as deps - skip in dry-run)
(Then mark collected state packages as explicit - skip in dry-run)
(Cleanup orphans - skip in dry-run)
Output summary
```

**Rationale**:
- Single iteration over packages
- Batch DB lookups minimize libalpm calls
- Clear error handling for missing packages
- Consistent with existing behavior

### 3. Batch local/sync DB lookup implementation

**Decision**: For local DB, iterate `localDB.PkgCache()` once and build a map. For sync DBs, iterate each repo's `PkgCache()`.

**Implementation**:
```go
// Build local package map in one pass
localPkgs := make(map[string]bool)
localDB.PkgCache().ForEach(func(pkg alpm.Package) error {
    localPkgs[pkg.Name()] = true
    return nil
})

// Similarly for each sync DB
for _, syncDB := range syncDBs {
    syncDB.PkgCache().ForEach(...)
}
```

**Rationale**:
- O(n) iteration where n = total packages in DB (not n queries)
- Single map construction, O(1) lookups per state package
- libalpm iterators are already lazy, no additional overhead

### 4. Dry-run behavior

**Decision**: Dry-run outputs exact packages that would be installed/removed without making any system changes.

**Implementation**:
- Skip `pacman -Syu` call
- Skip `pacman -D --asdeps` (mark all installed as deps)
- Skip `pacman -D --asexplicit` (mark state packages as explicit)
- Skip `pacman -Rns` orphan cleanup
- Still compute what WOULD happen for output

**Note on marking strategy**:
Instead of diffing between before/after installed packages, we simply:
1. After sync completes, run `pacman -D --asdeps` on ALL currently installed packages (this marks everything as deps)
2. Then run `pacman -D --asexplicit` on the collected state packages (this overrides them to explicit)

This is simpler and achieves the same result.

## Risks / Trade-offs

1. **[Risk]** dyalpm initialization requires root privileges
   - **[Mitigation]** This is same as pacman itself; if user can't run pacman, declpac won't work

2. **[Risk]** libalpm state becomes stale if another pacman instance runs concurrently
   - **[Mitigation]** Use proper locking, rely on pacman's own locking mechanism

3. **[Risk]** AUR packages still built sequentially
   - **[Acceptable]** Parallel AUR builds out of scope for this change

4. **[Risk]** Memory usage for large package lists
   - **[Mitigation]** Package map is ~100 bytes per package; 10k packages = ~1MB

## Migration Plan

1. Add `github.com/Jguer/dyalpm` to go.mod
2. Refactor `ValidatePackage()` to use dyalpm instead of subprocesses
3. Add `Installed bool` to `PackageInfo` struct
4. Implement new resolution algorithm in `categorizePackages()`
5. Update `Sync()` and `DryRun()` to use new algorithm
6. Test with various package combinations
7. Verify output matches previous behavior

## Open Questions

- **Q**: Should we also use dyalpm for `GetInstalledPackages()`?
- **A**: Yes, can use localDB.PkgCache().Collect() or iterate - aligns with overall approach