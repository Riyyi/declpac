## Context

Pacman, Arch Linux's package manager, lacks declarative package management and
doesn't support partial upgrades. Users managing system packages from scripts
or configuration files need a way to ensure their system state matches a declared
package list, with all packages at their latest available versions.

## Goals / Non-Goals

**Goals:**
- Provide a declarative interface for pacman package management through
  command-line interface
- Support flexible input sources (stdin and multiple state files) with
  additive merging
- **Automatically resolve transitive dependencies (users specify only direct packages)**
- Enforce full package upgrades (no partial upgrades) to prevent version
  mismatches
- Generate machine-readable output suitable for automated scripting
- Enable AI-driven implementation for new Go developers

**Non-Goals:**
- Package dependency resolution or smart upgrade scheduling
- Transaction rollback capabilities
- Pretty terminal output or interactive prompts
- Support for partial upgrades or selective package versions
- State persistence (tool only syncs, doesn't save state)

## Decisions

**Language: Go**
- Rationale: User wants modern CLI tool with native package management
  capabilities
- Go provides strong standard library, easy command-line parsing, and
  excellent for system-level operations
- User unfamiliar with Go - implementation will be AI-driven with detailed
  documentation

**Input Parsing Approach**
- Parse whitespace, tabs, newlines for package names
- Support multiple --state files: all additive including with stdin
- Validate package names against pacman available packages (optional validation)
- Empty state detection: print error to stderr and exit with code 1 (abort)

**State Merging Logic**
- All inputs are additive: combine all packages from stdin and all state files
- No conflict resolution: missing packages are added, duplicates accumulate
- Order-independent: inputs don't override each other

**Pacman Interaction**
- Use os/exec.Command to run pacman with -Sy (sync databases) and -S flag
- Run pacman -Syu with all target packages to ensure full upgrade
- Capture stderr/stdout for error reporting and machine output
- Detect pacman exit codes and translate to tool exit codes

**Error Handling Strategy**
- Parse pacman error messages from stderr output
- Distinguish between package-not-found (warning) vs. execution failures (errors)
- Return appropriate exit codes: 0 for success, non-zero for errors
- Include error details in output for scripting purposes

**AUR Integration**
- For packages not found in pacman repositories, check if available in AUR
- Use makepkg directly to build and install AUR packages (no AUR helpers)
- Clone AUR git repo to temp directory
- Run `makepkg -si` in temp directory for installation
- Capture stdout/stderr for output and error handling
- Report error to stderr if package not found in pacman or AUR

**Dependency Resolution**
- Use pacman's dependency resolution by passing all declared packages to `pacman -Syu`
- Pacman automatically resolves and includes transitive dependencies
- For AUR packages, makepkg handles dependency resolution during build
- No custom dependency resolution logic required - delegate to pacman/makepkg

**Explicit Marking & Orphan Cleanup**
- Before syncing, get list of all currently installed packages
- Mark declared state packages as explicitly installed via `pacman -D --explicit <pkg>`
- All other installed packages remain as non-explicit (dependencies)
- After sync completes, run `pacman -Rns $(pacman -Qdtq)` to remove orphaned packages
- Pacman -Rns $(pacman -Qdtq) only removes packages that are not explicitly
  installed and not required
- Capture and report number of packages removed during cleanup

**Package Version Management**
- Always force full system upgrade via `pacman -Syu <packages>`
- Pacman handles version selection automatically, ensuring latest versions
- No semantic version comparison or pinning logic required

## Risks / Trade-offs

**Known Risks:**
- [User unfamiliarity with Go] → Mitigation: Provide complete implementation
  with detailed comments; user can review and run without understanding deeply
- [Pacman error message parsing complexity] → Mitigation: Use broad error
  pattern matching; include error output as-is for debugging
- [Empty state triggers abort] → Mitigation: Print error to stderr and exit
  with code 1 instead of proceeding with empty state
- [Additive merging could accumulate duplicates] → Mitigation: Accept as
  design choice; pacman handles duplicates gracefully
- [Dependency conflicts could occur] → Mitigation: Let pacman handle standard
  conflicts; tool won't implement complex dependency resolution

**Trade-offs:**
- No conflict resolution: simpler merging but may include packages the system
  rejects
- Additive vs replacement: safer but less predictable for users expecting
  replacements
- No user prompts: fully automated for scripting but could be risky without
  warnings
