## ADDED Requirements

### Requirement: Package existence check via sync DBs

When resolving packages, the system SHALL first check sync databases (core, extra, multilib) to determine if a package exists in official repositories. If found in sync DBs, `Exists` SHALL be set to true.

#### Scenario: Package found in sync DB
- **WHEN** package name exists in any sync database
- **THEN** set `Exists: true`, `InAUR: false`

#### Scenario: Package not found in sync DB
- **WHEN** package name does not exist in any sync database
- **THEN** keep `Exists: false` for further lookup

### Requirement: Package installation check via local database

After checking sync DBs, the system SHALL check the local package database to determine if a package is installed. This is independent of existence check.

#### Scenario: Package installed locally
- **WHEN** package is installed on the system
- **THEN** set `Installed: true`

#### Scenario: Package not installed locally
- **WHEN** package is not installed on the system
- **THEN** set `Installed: false`

### Requirement: AUR fallback check

If package is not found in sync DBs, the system SHALL check the AUR as a fallback.

#### Scenario: Package found in AUR
- **WHEN** package exists in AUR but not in sync DBs
- **THEN** set `InAUR: true`

#### Scenario: Package not found anywhere
- **WHEN** package not in sync DBs, not in local DB, not in AUR
- **THEN** return error "package not found"

### Requirement: Validation at resolution end

After all checks complete, the system SHALL ensure every package has either `Exists: true` or `InAUR: true`. No package SHALL leave the resolver in an ambiguous state.

#### Scenario: All packages valid
- **WHEN** all packages resolved successfully
- **THEN** every package has Exists=true OR InAUR=true