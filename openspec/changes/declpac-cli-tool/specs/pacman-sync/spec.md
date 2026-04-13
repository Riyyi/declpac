## ADDED Requirements

### Requirement: Can sync system with declared package state

The system SHALL execute pacman operations to install and upgrade all declared packages.

#### Scenario: Full system upgrade with all packages
- **WHEN** system has declared package list
- **THEN** system shall run pacman -Syu with all declared package names

#### Scenario: No partial upgrades
- **WHEN** running pacman commands
- **THEN** system shall use -Syu flag (full system upgrade) ensuring all packages are latest

#### Scenario: Package availability check
- **WHEN** a package from input is not in pacman repositories
- **THEN** system shall report the error and include package name

#### Scenario: Pacman execution capture
- **WHEN** pacman commands execute
- **THEN** system shall capture both stdout and stderr for error reporting

#### Scenario: Exit code propagation
- **WHEN** pacman commands execute
- **THEN** system shall exit with code equal to pacman's exit code for success/failure detection