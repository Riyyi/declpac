## ADDED Requirements

### Requirement: Generate consistent output format

The system SHALL produce console output suitable for logging and scripting.

#### Scenario: Successful sync with changes
- **WHEN** sync completes with packages installed or removed
- **THEN** print to stdout: `Installed X packages, removed Y packages`
- **AND** exit with code 0

#### Scenario: No changes needed
- **WHEN** all packages already match declared state
- **THEN** print to stdout: `Installed 0 packages, removed 0 packages`
- **AND** exit with code 0

#### Scenario: Error during sync
- **WHEN** pacman or makepkg operation fails
- **THEN** print error to stderr with details
- **AND** exit with code 1

#### Scenario: Validation failure
- **WHEN** package validation fails (package not in pacman or AUR)
- **THEN** print error to stderr with package name
- **AND** exit with code 1

#### Scenario: Empty state input
- **WHEN** no packages provided from stdin or state files
- **THEN** print error to stderr: `empty state input`
- **AND** exit with code 1