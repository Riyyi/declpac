# Design: Refactor pkg Into Modular Packages

## New Structure

```
pkg/
├── pacman/     # write operations only
│   └── pacman.go
├── fetch/     # package resolution (NEW)
│   └── fetch.go
├── validation/ # DB freshness (existing, expand)
│   └── validation.go
├── output/    # (unchanged)
│   └── output.go
├── merge/    # (unchanged)
│   └── merge.go
└── input/    # (unchanged)
    └── input.go
```

## Package Responsibilities

### pkg/fetch (NEW)

```
type Fetcher struct {
    handle  dyalpm.Handle
    localDB dyalpm.Database
    syncDBs []dyalpm.Database
    aurCache map[string]AURPackage
}

func New() *Fetcher
func (f *Fetcher) Close() error
func (f *Fetcher) Resolve(packages []string) (map[string]*PackageInfo, error)
func (f *Fetcher) ListOrphans() ([]string, error)
```

Extracted from `pacman.go`:
- `buildLocalPkgMap()`
- `checkSyncDBs()`
- `resolvePackages()`
- AUR cache (`ensureAURCache()`, `fetchAURInfo()`)

### pkg/pacman (REFACTORED)

```
func Sync(packages []string) (*output.Result, error)
func DryRun(packages []string) (*output.Result, error)
func MarkAsExplicit(packages []string) error
func MarkAllAsDeps() error
func CleanupOrphans() (int, error)
```

Write actions only:
- `Sync()` - calls `Fetcher` for resolution, then pacman commands
- `DryRun()` - calls `Fetcher.Resolve()` + `ListOrphans()`
- `MarkAsExplicit()`, `MarkAllAsDeps()`
- `CleanupOrphans()` - calls `ListOrphans()` then removes

### pkg/validation (REFACTORED)

```
func CheckDBFreshness() error
```

Keep as-is: checks lock file age, auto-synces if stale.

Remove from `pacman.go`:
- `IsDBFresh()` - replaced by `CheckDBFreshness()`
- `SyncDB()` - called by validation when stale

### Orphan Deduplication

```go
// In fetch/fetch.go
func (f *Fetcher) ListOrphans() ([]string, error) {
    cmd := exec.Command("pacman", "-Qdtq")
    // ...
}

// In pacman/pacman.go
func CleanupOrphans() (int, error) {
    orphans, err := fetcher.ListOrphans()  // reuse
    if err != nil || len(orphans) == 0 {
        return 0, nil
    }
    // ... remove
}

func DryRun(...) (*output.Result, error) {
    orphans, err := fetcher.ListOrphans()  // reuse
    // ...
}
```

## Data Structures Move to fetch

```go
// In fetch/fetch.go
type PackageInfo struct {
    Name      string
    InAUR     bool
    Exists    bool
    Installed bool
    AURInfo   *AURPackage
    syncPkg   dyalpm.Package
}

type AURResponse struct {
    Results []AURPackage `json:"results"`
}

type AURPackage struct {
    Name        string `json:"Name"`
    PackageBase string `json:"PackageBase"`
    Version     string `json:"Version"`
    URL         string `json:"URL"`
}
```

## Import Changes

`pkg/pacman/fetch.go` will need:
- `github.com/Jguer/dyalpm`
- `github.com/Riyyi/declpac/pkg/output`

`pkg/pacman/pacman.go` will need:
- `github.com/Riyyi/declpac/pkg/fetch`
- `github.com/Riyyi/declpac/pkg/output`

## Dependencies to Add

- New import: `pkg/fetch` in `pacman` package

## No Changes To

- CLI entry points
- `output.Result` struct
- `input.ReadPackages()`
- `merge.Merge()`