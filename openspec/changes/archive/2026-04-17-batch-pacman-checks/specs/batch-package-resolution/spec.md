## ADDED Requirements

### Requirement: Batch package resolution from local, sync, and AUR databases
The system SHALL resolve packages in a single pass through local DB → sync DBs → AUR using batch operations to minimize subprocess/API calls.

#### Scenario: Package exists in local DB
- **WHEN** a package from collected state exists in the local database
- **THEN** the system SHALL mark it as found, set `Installed=true`, and exclude it from AUR queries

#### Scenario: Package exists in sync DB
- **WHEN** a package from collected state does NOT exist in local DB but exists in ANY enabled sync database
- **THEN** the system SHALL mark it as found, set `Installed=false`, and exclude it from AUR queries

#### Scenario: Package exists only in AUR
- **WHEN** a package from collected state does NOT exist in local or sync databases but exists in AUR
- **THEN** the system SHALL mark it as found with `InAUR=true`, set `Installed=false`, and use the cached AUR info

#### Scenario: Package not found anywhere
- **WHEN** a package from collected state is NOT in local DB, NOT in any sync DB, and NOT in AUR
- **THEN** the system SHALL return an error listing the unfound package(s)

#### Scenario: Batch AUR query
- **WHEN** multiple packages need AUR lookup
- **THEN** the system SHALL make a SINGLE HTTP request to AUR RPC with all package names (existing behavior preserved)

### Requirement: Efficient local DB lookup using dyalpm
The system SHALL use dyalpm's `PkgCache()` iterator to build a lookup map in O(n) time, where n is total packages in local DB, instead of O(n*m) subprocess calls.

#### Scenario: Build local package map
- **WHEN** initializing package resolution
- **THEN** the system SHALL iterate localDB.PkgCache() once and store all package names in a map for O(1) lookups

#### Scenario: Check package in local map
- **WHEN** checking if a package exists in local DB
- **THEN** the system SHALL perform an O(1) map lookup instead of spawning a subprocess

### Requirement: Efficient sync DB lookup using dyalpm
The system SHALL use each sync DB's `PkgCache()` iterator to check packages across all enabled repositories.

#### Scenario: Check package in sync DBs
- **WHEN** a package is not found in local DB
- **THEN** the system SHALL check all enabled sync databases using their iterators

#### Scenario: Package found in multiple sync repos
- **WHEN** a package exists in more than one sync repository (e.g., core and community)
- **THEN** the system SHALL use the first match found

### Requirement: Track installed status in PackageInfo
The system SHALL include an `Installed bool` field in `PackageInfo` to indicate whether the package is currently installed.

#### Scenario: Package is installed
- **WHEN** a package exists in the local database
- **THEN** `PackageInfo.Installed` SHALL be `true`

#### Scenario: Package is not installed
- **WHEN** a package exists only in sync DB or AUR (not in local DB)
- **THEN** `PackageInfo.Installed` SHALL be `false`

### Requirement: Mark installed packages as deps, then state packages as explicit
After package sync completes, the system SHALL mark all installed packages as dependencies, then override the collected state packages to be explicit. This avoids diffing before/after states.

#### Scenario: Mark all installed as deps
- **WHEN** package sync has completed (non-dry-run)
- **THEN** the system SHALL run `pacman -D --asdeps` to mark ALL currently installed packages as dependencies

#### Scenario: Override state packages to explicit
- **WHEN** all installed packages have been marked as deps
- **THEN** the system SHALL run `pacman -D --asexplicit` on the collected state packages, overriding their dependency status

#### Scenario: Dry-run skips marking
- **WHEN** operating in dry-run mode
- **THEN** the system SHALL NOT execute any `pacman -D` marking operations