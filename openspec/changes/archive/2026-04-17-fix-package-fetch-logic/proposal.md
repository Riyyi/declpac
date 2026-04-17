## Why

The `Resolve` function in `pkg/fetch/fetch.go` has incorrect logic flow. It
initializes all packages with `Exists: true`, then checks local DB first, then
sync DBs, then AUR. This wrong order causes incorrect package state
classification - packages that exist only in AUR may be incorrectly marked as
found.

## What Changes

- Fix initialization: packages should start with `Exists: false` (unknown), not `true`
- Fix order: check sync DBs BEFORE local DB to determine availability
- Separate independent concerns: "available" (sync/AUR) from "installed" (local)
- All packages must validate: either `Exists: true` or `InAUR: true`

## Capabilities

### New Capabilities
- `package-resolve-logic`: Correct resolution algorithm for pacman packages

### Modified Capabilities
- None

## Impact

- `pkg/fetch/fetch.go`: `Resolve` function
