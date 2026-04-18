## Why

The tool exits without saving the full pacman output, making debugging difficult
when operations fail. Users need a persistent log of all pacman operations for
debugging and audit.

## What Changes

- Add state directory initialization creating `~/.local/state/declpac` if not exists
- Open/manage a single log file at `/var/log/declpac.log`
- Instrument all state-modifying exec calls in `pkg/pacman/pacman.go` to tee or append output to this file
- Skip debug messages (internal timing logs)
- Capture and write errors before returning

## Capabilities

### New Capabilities
- Operation logging: Persist stdout/stderr from all pacman operations

### Modified Capabilities
- None

## Impact

- `pkg/pacman/pacman.go`: Instrument all state-modifying functions to write to log file
- New module: May create `pkg/state/state.go` or similar for log file management
