## ADDED Requirements

### Requirement: Can detect empty state input

The system SHALL detect when no packages are provided in any input source.

#### Scenario: Empty state warning
- **WHEN** no package names are found in stdin or state files
- **THEN** system shall print a warning to stderr

#### Scenario: Empty state warning message
- **WHEN** warning is printed for empty state
- **THEN** message shall say: "Called without state, aborting.."

#### Scenario: Abort on empty state
- **WHEN** empty state is detected
- **THEN** system shall exit with code 1 (error: no packages to sync)
