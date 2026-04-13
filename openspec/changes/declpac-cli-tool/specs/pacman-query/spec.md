## ADDED Requirements

### Requirement: Query pacman package database via libalpm

The system SHALL use libalpm for fast local queries against the pacman
package database.

#### Scenario: Query installed packages
- **WHEN** needing list of currently installed packages
- **THEN** query via libalpm alpm_list_get(ALPM_LIST_PACKAGES)

#### Scenario: Query available packages
- **WHEN** validating package exists in pacman repos
- **THEN** query via libalpm alpm_db_get_pkg()

#### Scenario: Check database last sync time
- **WHEN** checking if database needs refresh
- **THEN** check /var/lib/pacman/db.lock timestamp

#### Scenario: Query foreign packages
- **WHEN** checking if package is AUR-installed
- **THEN** query via libalpm alpm_db_get_pkg(ALPM_DB_TYPE_LOCAL)