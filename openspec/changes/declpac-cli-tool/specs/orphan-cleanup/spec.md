## ADDED Requirements

### Requirement: Can mark non-state packages as non-explicit

The system SHALL mark all packages not declared in the state as non-explicit
(dependencies) before syncing, so they can be safely removed later.

#### Scenario: Mark packages as non-explicit
- **WHEN** packages are declared in state
- **THEN** system shall mark those packages as explicitly installed via pacman -D --explicit
- **AND** mark all other installed packages as non-explicit

### Requirement: Can clean up orphaned packages

After syncing, the system SHALL remove packages that are no longer required
(as dependencies of removed packages).

#### Scenario: Orphan cleanup after sync
- **WHEN** sync operation completes successfully
- **THEN** system shall run pacman -Rns to remove unneeded dependencies
- **AND** report the number of packages removed

#### Scenario: Orphan cleanup respects explicitly installed
- **WHEN** a package not in state is marked as explicitly installed by user
- **THEN** system shall NOT remove it during orphan cleanup

#### Scenario: No orphans to clean
- **WHEN** there are no orphaned packages to remove
- **THEN** system shall report "No packages to remove" in output