## ADDED Requirements

### Requirement: Can load package states from state files

The system SHALL load package lists from state files specified via --state flag.

#### Scenario: Single state file loading
- **WHEN** a state file is provided via --state flag
- **THEN** system shall read package names from the file content

#### Scenario: Multiple state file loading
- **WHEN** multiple --state flags are provided
- **THEN** system shall load package names from each file

#### Scenario: State file format is text list
- **WHEN** loading from state files
- **THEN** system shall parse whitespace-delimited package names

#### Scenario: File path validation
- **WHEN** a state file path points to a non-existent file
- **THEN** system shall report a file read error