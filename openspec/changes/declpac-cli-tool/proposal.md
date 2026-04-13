## Why

Pacman doesn't support declarative package management or partial upgrades,
making it challenging to maintain consistent system states. This tool provides
a clean, scriptable interface to synchronize the system with a declared package
list, ensuring all packages are at the latest version.

## What Changes

- Add CLI tool for declarative pacman package management
- Support stdin input for package lists
- Support multiple --state file inputs
- Merge all inputs (additive strategy with empty state warning)
- Force full upgrade of all packages (no partial upgrades)
- Handle transitive dependencies automatically (users only specify direct packages)
- Mark non-state packages as non-explicit before sync
- After sync, remove orphaned packages (cleanup)
- Support AUR packages: try pacman first, then makepkg, then report errors
- Machine-readable output (install/remove counts, exit codes)
- No conflict resolution for missing packages (append-only)
- Print error to stderr for empty state input and exit with code 1

## Capabilities

### New Capabilities

- **stdin-input**: Read package lists from standard input stream
- **state-files**: Load declarative package states from files
- **input-merging**: Combine multiple input sources (stdin + multiple --state files)
- **state-validation**: Validate package names and empty states
- **transitive-deps**: Automatically resolve all transitive dependencies
- **pacman-sync**: Execute pacman operations to match declared state
- **aur-sync**: Handle AUR packages: pacman first, then fall back to makepkg,
                  then report errors
- **machine-output**: Generate machine-readable sync results for scripting

### Modified Capabilities

None (new tool)