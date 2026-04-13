## ADDED Requirements

### Requirement: Can resolve transitive dependencies automatically

The system SHALL automatically determine and install all transitive dependencies
for declared packages, so users only need to specify direct packages.

#### Scenario: Transitive dependencies are resolved
- **WHEN** a package with dependencies is declared in state
- **THEN** system shall identify all transitive dependencies via pacman
- **AND** include them in the installation operation

#### Scenario: Dependency resolution includes AUR dependencies
- **WHEN** a package's transitive dependencies include AUR packages
- **THEN** system shall resolve those AUR dependencies via makepkg

#### Scenario: Shared dependencies are deduplicated
- **WHEN** multiple declared packages share dependencies
- **THEN** system shall install each shared dependency only once

#### Scenario: Missing dependency handling
- **WHEN** a declared package has a dependency not available in pacman or AUR
- **THEN** system shall report error for the missing dependency