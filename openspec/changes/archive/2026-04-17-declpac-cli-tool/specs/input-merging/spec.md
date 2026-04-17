## ADDED Requirements

### Requirement: Can merge multiple input sources

The system SHALL combine package lists from all inputs using an additive
strategy without conflict resolution.

#### Scenario: Additive merging of all inputs
- **WHEN** stdin and multiple state files provide package lists
- **THEN** system shall include all packages from all inputs in the final state

#### Scenario: Input priority is last-writer-wins per file
- **WHEN** multiple state files contain the same package name
- **THEN** the package from the last provided file in command-line order takes
  precedence

#### Scenario: Missing packages from stdin are added
- **WHEN** stdin contains packages not in state files
- **THEN** those packages shall be added to the final state

#### Scenario: Duplicate packages are deduplicated
- **WHEN** the same package appears in multiple inputs
- **THEN** it shall be included once in the final state (map deduplication)
