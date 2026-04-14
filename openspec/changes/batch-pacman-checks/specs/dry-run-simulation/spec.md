## ADDED Requirements

### Requirement: Dry-run shows packages to install without making changes
In dry-run mode, the system SHALL compute what WOULD happen without executing any pacman operations.

#### Scenario: Dry-run lists packages to install
- **WHEN** dry-run is enabled and packages need to be installed
- **THEN** the system SHALL populate `Result.ToInstall` with all packages that would be installed (both sync and AUR)

#### Scenario: Dry-run lists packages to remove
- **WHEN** dry-run is enabled and orphan packages exist
- **THEN** the system SHALL NOT calculate or populate `Result.ToRemove` - orphan detection is skipped entirely in dry-run mode

#### Scenario: Dry-run skips pacman sync
- **WHEN** dry-run is enabled
- **THEN** the system SHALL NOT execute `pacman -Syu` for package installation

#### Scenario: Dry-run skips explicit/deps marking
- **WHEN** dry-run is enabled
- **THEN** the system SHALL NOT execute `pacman -D --asdeps` or `pacman -D --asexplicit`

#### Scenario: Dry-run skips orphan cleanup
- **WHEN** dry-run is enabled
- **THEN** the system SHALL NOT execute `pacman -Rns` for orphan removal

#### Scenario: Dry-run outputs count summary
- **WHEN** dry-run is enabled
- **THEN** the system SHALL still compute and output `Result.Installed` and `Result.Removed` counts as if the operations had run