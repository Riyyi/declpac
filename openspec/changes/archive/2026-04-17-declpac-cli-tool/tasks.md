## 1. Project Setup

- [x] 1.1 Initialize Go module with proper imports
- [x] 1.2 Add required dependencies (dyalpm wrapper, Jguer/aur)
- [x] 1.3 Set up project structure (cmd/declpac/main.go, pkg/ subdirectory)
- [x] 1.4 Add libalpm initialization and handle

## 2. Input Parsing

- [x] 2.1 Implement stdin reader to collect package names
- [x] 2.2 Implement state file reader for text-list format
- [x] 2.3 Add whitespace normalization for package names
- [x] 2.4 Create package name set data structure

## 3. Input Merging

- [x] 3.1 Implement additive merging of stdin and state file packages
- [x] 3.2 Handle multiple --state flags with last-writer-wins per file
- [x] 3.3 Implement duplicate package handling (no deduplication)

## 4. State Validation

- [x] 4.1 Implement empty state detection (no packages found)
- [x] 4.2 Add stderr error output for empty state
- [x] 4.3 Set exit code 1 for empty state case (abort, not proceed)
- [x] 4.4 Check pacman DB freshness (db.lock timestamp)
- [x] 4.5 Run pacman -Syy if DB older than 1 day
- [x] 4.6 Validate packages via libalpm (pacman repos)
- [x] 4.7 Validate packages via Jguer/aur (AUR)
- [x] 4.8 Fail fast with error if package not found

## 5. Pacman Integration (Hybrid: query via libalpm, modify via exec)

- [x] 5.1 Initialize libalpm handle for queries
- [x] 5.2 Implement libalpm query for installed packages
- [x] 5.3 Implement libalpm query for available packages
- [x] 5.4 Implement pacman -Syy command execution (DB refresh)
- [x] 5.5 Implement pacman -Syu command execution wrapper
- [x] 5.6 Add command-line argument construction with package list
- [x] 5.7 Capture pacman stdout and stderr output
- [x] 5.8 Implement pacman error message parsing
- [x] 5.9 Handle pacman exit codes for success/failure detection
- [x] 5.10 Verify pacman automatically resolves transitive dependencies

## 6. Explicit Marking & Orphan Cleanup

- [x] 6.1 Get list of currently installed packages before sync
- [x] 6.2 Mark declared state packages as explicitly installed via pacman -D --explicit
- [x] 6.3 Run pacman sync operation (5.x series)
- [x] 6.4 Run pacman -Rns to remove orphaned packages
- [x] 6.5 Capture and report number of packages removed
- [x] 6.6 Handle case where no orphans exist (no packages removed)

## 7. AUR Integration

- [x] 7.1 Implement AUR package lookup via Jguer/aur library
- [x] 7.2 Check package not in pacman repos first (via libalpm)
- [x] 7.3 Query AUR for missing packages
- [x] 7.4 Implement AUR fallback using makepkg (direct build, not AUR helper)
- [x] 7.5 Clone AUR package git repo to temp directory
- [x] 7.6 Run makepkg -si in temp directory for installation
- [x] 7.7 Upgrade existing AUR packages to latest (makepkg rebuild)
- [x] 7.8 Add stderr error reporting for packages not in pacman or AUR
- [x] 7.9 Capture makepkg stdout and stderr for output parsing
- [x] 7.10 Handle makepkg exit codes for success/failure detection

## 8. Output Generation

- [x] 8.1 Parse pacman output for installed package count
- [x] 8.2 Parse pacman output for removed package count (orphan cleanup)
- [x] 8.3 Generate output: `Installed X packages, removed Y packages`
- [x] 8.4 Handle 0 packages case: `Installed 0 packages, removed 0 packages`
- [x] 8.5 Print errors to stderr
- [x] 8.6 Set exit code 0 for success, 1 for errors

## 9. CLI Interface

- [x] 9.1 Implement --state flag argument parsing
- [x] 9.2 Implement stdin input handling from /dev/stdin
- [x] 9.3 Set up correct CLI usage/help message
- [x] 9.4 Implement flag order validation

## 10. Integration & Testing

- [x] 10.1 Wire together stdin -> state files -> merging -> validation -> pacman sync -> orphan cleanup -> output
- [x] 10.2 Test empty state error output and exit code 1
- [x] 10.3 Test single state file parsing
- [x] 10.4 Test multiple state file merging
- [x] 10.5 Test stdin input parsing
- [x] 10.6 Test explicit marking before sync
- [x] 10.7 Test pacman command execution with real packages
- [x] 10.8 Test orphan cleanup removes unneeded packages
- [x] 10.9 Test AUR fallback with makepkg for AUR package
- [x] 10.10 Test error handling for missing packages
- [x] 10.11 Generate final binary

## 11. Dry-Run Mode

- [x] 11.1 Add --dry-run flag to CLI argument parsing
- [x] 11.2 Implement DryRun function to query current state
- [x] 11.3 Compare declared packages to current installations
- [x] 11.4 Identify packages to install (not currently installed)
- [x] 11.5 Identify orphans to remove via pacman -Qdtq
- [x] 11.6 Output "Would install:" and "Would remove:" sections
