## 1. Project Setup

- [ ] 1.1 Initialize Go module with proper imports
- [ ] 1.2 Add required dependencies (dyalpm wrapper, Jguer/aur)
- [ ] 1.3 Set up project structure (cmd/declpac/main.go, pkg/ subdirectory)
- [ ] 1.4 Add libalpm initialization and handle

## 2. Input Parsing

- [ ] 2.1 Implement stdin reader to collect package names
- [ ] 2.2 Implement state file reader for text-list format
- [ ] 2.3 Add whitespace normalization for package names
- [ ] 2.4 Create package name set data structure

## 3. Input Merging

- [ ] 3.1 Implement additive merging of stdin and state file packages
- [ ] 3.2 Handle multiple --state flags with last-writer-wins per file
- [ ] 3.3 Implement duplicate package handling (no deduplication)

## 4. State Validation

- [ ] 4.1 Implement empty state detection (no packages found)
- [ ] 4.2 Add stderr error output for empty state
- [ ] 4.3 Set exit code 1 for empty state case (abort, not proceed)
- [ ] 4.4 Check pacman DB freshness (db.lock timestamp)
- [ ] 4.5 Run pacman -Syy if DB older than 1 day
- [ ] 4.6 Validate packages via libalpm (pacman repos)
- [ ] 4.7 Validate packages via Jguer/aur (AUR)
- [ ] 4.8 Fail fast with error if package not found

## 5. Pacman Integration (Hybrid: query via libalpm, modify via exec)

- [ ] 5.1 Initialize libalpm handle for queries
- [ ] 5.2 Implement libalpm query for installed packages
- [ ] 5.3 Implement libalpm query for available packages
- [ ] 5.4 Implement pacman -Syy command execution (DB refresh)
- [ ] 5.5 Implement pacman -Syu command execution wrapper
- [ ] 5.6 Add command-line argument construction with package list
- [ ] 5.7 Capture pacman stdout and stderr output
- [ ] 5.8 Implement pacman error message parsing
- [ ] 5.9 Handle pacman exit codes for success/failure detection
- [ ] 5.10 Verify pacman automatically resolves transitive dependencies

## 6. Explicit Marking & Orphan Cleanup

- [ ] 6.1 Get list of currently installed packages before sync
- [ ] 6.2 Mark declared state packages as explicitly installed via pacman -D --explicit
- [ ] 6.3 Run pacman sync operation (5.x series)
- [ ] 6.4 Run pacman -Rsu to remove orphaned packages
- [ ] 6.5 Capture and report number of packages removed
- [ ] 6.6 Handle case where no orphans exist (no packages removed)

## 7. AUR Integration

- [ ] 7.1 Implement AUR package lookup via Jguer/aur library
- [ ] 7.2 Check package not in pacman repos first (via libalpm)
- [ ] 7.3 Query AUR for missing packages
- [ ] 7.4 Implement AUR fallback using makepkg (direct build, not AUR helper)
- [ ] 7.5 Clone AUR package git repo to temp directory
- [ ] 7.6 Run makepkg -si in temp directory for installation
- [ ] 7.7 Upgrade existing AUR packages to latest (makepkg rebuild)
- [ ] 7.8 Add stderr error reporting for packages not in pacman or AUR
- [ ] 7.9 Capture makepkg stdout and stderr for output parsing
- [ ] 7.10 Handle makepkg exit codes for success/failure detection

## 8. Output Generation

- [ ] 8.1 Parse pacman output for installed package count
- [ ] 8.2 Parse pacman output for removed package count (orphan cleanup)
- [ ] 8.3 Generate output: `Installed X packages, removed Y packages`
- [ ] 8.4 Handle 0 packages case: `Installed 0 packages, removed 0 packages`
- [ ] 8.5 Print errors to stderr
- [ ] 8.6 Set exit code 0 for success, 1 for errors

## 9. CLI Interface

- [ ] 9.1 Implement --state flag argument parsing
- [ ] 9.2 Implement stdin input handling from /dev/stdin
- [ ] 9.3 Set up correct CLI usage/help message
- [ ] 9.4 Implement flag order validation

## 10. Integration & Testing

- [ ] 10.1 Wire together stdin -> state files -> merging -> validation -> pacman sync -> orphan cleanup -> output
- [ ] 10.2 Test empty state error output and exit code 1
- [ ] 10.3 Test single state file parsing
- [ ] 10.4 Test multiple state file merging
- [ ] 10.5 Test stdin input parsing
- [ ] 10.6 Test explicit marking before sync
- [ ] 10.7 Test pacman command execution with real packages
- [ ] 10.8 Test orphan cleanup removes unneeded packages
- [ ] 10.9 Test AUR fallback with makepkg for AUR package
- [ ] 10.10 Test error handling for missing packages
- [ ] 10.11 Generate final binary
