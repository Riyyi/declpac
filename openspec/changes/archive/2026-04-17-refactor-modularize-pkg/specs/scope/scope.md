# Scope

## In Scope

- Extract package resolution from pacman.go to pkg/fetch
- Deduplicate orphan listing
- Keep pacman write operations in pacman package
- Maintain existing CLI API

## Out of Scope

- New features
- New package management backends (e.g., libalpm alternatives)
- Config file changes
- State file format changes