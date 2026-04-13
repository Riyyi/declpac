## ADDED Requirements

### Requirement: Can handle AUR packages with pacman fallback

The system SHALL attempt to install AUR packages by first trying pacman
repositories, then falling back to AUR when pacman fails, and finally reporting
errors for packages still missing.

#### Scenario: Install from pacman first
- **WHEN** package is in pacman repositories
- **THEN** system shall install via pacman

#### Scenario: Fall back to AUR
- **WHEN** package is not in pacman repositories but is in AUR
- **THEN** system shall attempt installation via makepkg (direct AUR build)

#### Scenario: Report error for missing packages
- **WHEN** package is not in pacman repositories or AUR
- **THEN** system shall report error and print to stderr

#### Scenario: Continue for remaining packages
- **WHEN** some packages are installed and some fail
- **THEN** system shall continue installation process for successful packages
