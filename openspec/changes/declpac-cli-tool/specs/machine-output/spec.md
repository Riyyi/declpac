## ADDED Requirements

### Requirement: Can generate machine-readable output

The system SHALL produce output suitable for automated scripting.

#### Scenario: Install/remove count output
- **WHEN** sync completes successfully
- **THEN** system shall output: "<installed_count> package(s) installed, <removed_count> package(s) removed"

#### Scenario: Empty output format
- **WHEN** no packages to sync
- **THEN** system shall output nothing (or indicate no changes)

#### Scenario: Error details in output
- **WHEN** pacman operation fails
- **THEN** system shall include error details in output (package names, pacman error messages)

#### Scenario: Exit code for scripting
- **WHEN** sync completes
- **THEN** system shall return exit code 0 for success, non-zero for errors