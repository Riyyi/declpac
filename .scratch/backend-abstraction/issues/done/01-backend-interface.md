# 01 - Define Backend Interface and PackageInfo

Status: done

## Description

Create `pkg/backend/backend.go` with the `Backend` interface, `PackageInfo` struct, and `New()` auto-detection function.

### Backend interface

```go
type Backend interface {
    Name() string
    Resolve(packages []string) ([]PackageInfo, error)
    List() ([]string, error)
    ListExplicit() ([]string, error)
    ListOrphans() ([]string, error)
    RefreshDB(logWriter io.Writer) error
    InstallPackages(pkgs []PackageInfo, logWriter io.Writer) error
    RemoveOrphans(orphans []string, logWriter io.Writer) (int, error)
    MarkAs(packages []string, flag string, logWriter io.Writer) error
    Prune(logWriter io.Writer) (int, error)
}
```

### PackageInfo struct

```go
type PackageInfo struct {
    Name      string
    Source    string // "repo", "aur", "copr"
    Available bool
    Installed bool
    Provided  string
}
```

### New() function

Uses `exec.LookPath` to detect `pacman` or `dnf`. Returns clean error if both or neither found. Returns the concrete `*pacman.Pacman` or `*dnf.Dnf` behind the interface.

## Acceptance Criteria

- `pkg/backend/backend.go` compiles
- `New()` detects pacman when only pacman is on PATH
- `New()` detects dnf when only dnf is on PATH
- `New()` returns error when both are found
- `New()` returns error when neither is found