## ADDED Requirements

### Requirement: Handle AUR packages with fallback and upgrade

The system SHALL first attempt to install packages via pacman, then fall back
to AUR for packages not found in repos, and upgrade AUR packages to
latest versions.

#### Scenario: Try pacman first
- **WHEN** package is in pacman repositories
- **THEN** install via pacman -Syu

#### Scenario: Fall back to AUR
- **WHEN** package is not in pacman repositories but is in AUR
- **THEN** query AUR via Jguer/aur library
- **AND** build and install with makepkg -si

#### Scenario: Upgrade AUR packages
- **WHEN** AUR package is already installed but outdated
- **THEN** rebuild and reinstall with makepkg to get latest version

#### Scenario: Report error for missing packages
- **WHEN** package is not in pacman repositories or AUR
- **THEN** print error to stderr with package name
- **AND** exit with code 1

#### Scenario: AUR build failure
- **WHEN** makepkg fails to build package
- **THEN** print makepkg error to stderr
- **AND** exit with code 1