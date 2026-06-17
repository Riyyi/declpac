# 03 - Extract Shared Orchestration to pkg/backend

Status: ready-for-agent

## Description

Move DryRun and Sync orchestration logic from `pkg/pacman/` and `pkg/pacman/read/` into `pkg/backend/`. Create `dryrun.go` and `sync.go` that call Backend interface methods.

### dryrun.go

Extract `read.DryRun` logic into a shared function:
- Takes `Backend`, packages, config
- Calls `backend.Resolve()`, `backend.ListOrphans()`
- Categorizes into toInstall/toRemove
- Returns `*output.Result`

### sync.go

Extract `pacman.Sync` logic into a shared function:
- Takes `Backend`, packages, config
- Runs safety check (NoCheck guard) using `backend.ListExplicit()`
- Calls `backend.RefreshDB()` if needed (handled inside RefreshDB now)
- Calls `backend.Resolve()`, categorizes
- Calls `backend.InstallPackages()` with categorized packages
- Runs prune sanity check using `backend.List()`
- Calls `backend.Prune()` if config.Prune
- Returns `*output.Result`

### main.go update

- Call `backend.New()` to get a Backend
- Call shared `backend.DryRun(backend, packages, config)` or `backend.Sync(backend, packages, config)` instead of `read.DryRun` / `pacman.Sync`

## Acceptance Criteria

- `pkg/backend/dryrun.go` and `pkg/backend/sync.go` exist and compile
- `cmd/declpac/main.go` uses `backend.New()` and shared orchestration functions
- `pacman.Sync` and `read.DryRun` are removed (all logic moved)
- Existing behavior unchanged when running with pacman backend