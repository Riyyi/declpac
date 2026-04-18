## Tasks

- [x] 1. Create state module

Create `pkg/state/state.go`:
- `OpenLog()` - opens `/var/log/declpac.log` in append mode
- `GetLogWriter()` - returns the raw log file writer (for MultiWriter)
- `Write(msg []byte)` - writes message with timestamp + dashes separator
- `Close()` - closes the file

- [x] 2. Wire into main.go

In `cmd/declpac/main.go` `run()`:
- Call `OpenLog()` at start
- `defer Close()` log

- [x] 3. Instrument pkg/pacman

Modify `pkg/pacman/pacman.go`:

Each state-modifying function writes timestamp ONCE at start, then captures output:
- Write `timestamp - operation name` to log
- Run command, capture output
- Write captured output to log
- Write output to terminal

**Functions updated:**
- `SyncPackages()` - write output with timestamp
- `CleanupOrphans()` - write output with timestamp
- `MarkAllAsDeps()` - write operation name with timestamp before running
- `MarkAsExplicit()` - write operation name with timestamp before running
- `InstallAUR()` - write "Cloning ..." and "Building package..." with timestamps
- Error handling - `state.Write([]byte("error: ..."))` for all error paths

- [x] 4. Add io import

Add `io` to imports in pacman.go

- [x] 5. Test

Run a sync operation and verify log file created at `/var/log/declpac.log`
