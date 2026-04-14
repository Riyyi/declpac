## Why

The program fails to find packages that exist in official repositories (like `cmake` in `extra`). The dyalpm library's `SyncDBs()` returns an empty list, so the code never searches the sync databases and falls through to checking the AUR, which also doesn't have the package.

## What Changes

- Register sync databases manually after initializing the dyalpm handle
- Loop through known repos (core, extra, multilib) and call `RegisterSyncDB` for each
- Handle the case where a repo might not be configured in pacman.conf

## Capabilities

### New Capabilities
- None - this is a bug fix to existing functionality

### Modified Capabilities
- None

## Impact

- `pkg/pacman/pacman.go`: Modify `New()` function to register sync DBs after getting them from the handle