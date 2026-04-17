## ADDED Requirements

### Requirement: Validate package names exist in pacman or AUR

The system SHALL validate that all declared packages exist in pacman
repositories or AUR before attempting to sync.

#### Scenario: Empty state detection
- **WHEN** no package names are found in stdin or state files
- **THEN** print error to stderr: `empty state input`
- **AND** exit with code 1

#### Scenario: Validate package in pacman repos
- **WHEN** package exists in pacman repositories
- **THEN** validation passes

#### Scenario: Validate package in AUR
- **WHEN** package not in pacman repos but exists in AUR
- **THEN** validation passes

#### Scenario: Package not found
- **WHEN** package not in pacman repos or AUR
- **THEN** print error to stderr with package name
- **AND** exit with code 1

#### Scenario: Database freshness check
- **WHEN** pacman database last sync was more than 1 day ago
- **THEN** run pacman -Syy to refresh before validation